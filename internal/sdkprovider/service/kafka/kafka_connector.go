package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenKafkaConnectorSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"connector_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The kafka connector name.").ForceNew().Build(),
	},
	"config": {
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "The Kafka Connector configuration parameters.",
	},
	"plugin_author": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Kafka connector author.",
	},
	"plugin_class": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Kafka connector Java class.",
	},
	"plugin_doc_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Kafka connector documentation URL.",
	},
	"plugin_title": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Kafka connector title.",
	},
	"plugin_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Kafka connector type.",
	},
	"plugin_version": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The version of the kafka connector.",
	},
	"task": {
		Type:        schema.TypeSet,
		Description: "List of tasks of a connector.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connector": {
					Type:        schema.TypeString,
					Description: "The name of the related connector.",
					Computed:    true,
				},
				"task": {
					Type:        schema.TypeInt,
					Description: "The task id of the task.",
					Computed:    true,
				},
			},
		},
	},
}

func ResourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka connectors resource allows the creation and management of Aiven Kafka connectors.",
		CreateContext: resourceKafkaConnectorCreate,
		ReadContext:   resourceKafkaConnectorRead,
		UpdateContext: resourceKafkaTConnectorUpdate,
		DeleteContext: resourceKafkaConnectorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaConnectorSchema,
		CustomizeDiff: customdiff.IfValueChange("config",
			kafkaConnectorConfigNameShouldNotBeEmpty(),
			customizeDiffKafkaConnectorConfigName(),
		),
	}
}

// customizeDiffKafkaConnectorConfigName `config.name` should be equal to `connector_name`
func customizeDiffKafkaConnectorConfigName() func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
	return func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		connectorName := diff.Get("connector_name").(string)
		config := make(aiven.KafkaConnectorConfig)
		for k, cS := range diff.Get("config").(map[string]interface{}) {
			config[k] = cS.(string)
		}

		if connectorName != config["name"] {
			return fmt.Errorf("config.name should be equal to the connector_name")
		}
		return nil
	}
}

// kafkaConnectorConfigNameShouldNotBeEmpty `config.name` should not be empty
func kafkaConnectorConfigNameShouldNotBeEmpty() func(ctx context.Context, oldValue interface{}, newValue interface{}, meta interface{}) bool {
	return func(ctx context.Context, oldValue, newValue, meta interface{}) bool {
		for k, cS := range newValue.(map[string]interface{}) {
			if k == "name" && cS.(string) != "" {
				return true
			}
		}
		return false
	}
}

func flattenKafkaConnectorTasks(r *aiven.KafkaConnector) []map[string]interface{} {
	var tasks []map[string]interface{}

	for _, taskS := range r.Tasks {
		task := map[string]interface{}{
			"connector": taskS.Connector,
			"task":      taskS.Task,
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func resourceKafkaConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, connectorName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{"IN_PROGRESS"},
		Target:  []string{"OK"},
		Refresh: func() (interface{}, string, error) {
			list, err := m.(*aiven.Client).KafkaConnectors.List(project, serviceName)
			if err != nil {
				log.Printf("[DEBUG] Kafka Connectors list waiter err %s", err.Error())
				if aiven.IsNotFound(err) {
					return nil, "", err
				}

				return nil, "IN_PROGRESS", nil
			}

			return list, "OK", nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutRead),
		MinTimeout: 2 * time.Second,
	}
	res, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	var found bool
	for _, r := range res.(*aiven.KafkaConnectorsResponse).Connectors {
		if r.Name == connectorName {
			found = true
			if err := d.Set("project", project); err != nil {
				return diag.Errorf("error setting Kafka Connector `project` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("service_name", serviceName); err != nil {
				return diag.Errorf("error setting Kafka Connector `service_name` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("connector_name", connectorName); err != nil {
				return diag.Errorf("error setting Kafka Connector `connector_name` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("config", r.Config); err != nil {
				return diag.Errorf("error setting Kafka Connector `config` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_author", r.Plugin.Author); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_author` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_class", r.Plugin.Class); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_class` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_doc_url", r.Plugin.DocumentationURL); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_doc_url` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_title", r.Plugin.Title); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_title` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_type", r.Plugin.Type); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_type` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_version", r.Plugin.Version); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_version` for resource %s: %s", d.Id(), err)
			}

			tasks := flattenKafkaConnectorTasks(&r)
			if err := d.Set("task", tasks); err != nil {
				return diag.Errorf("error setting Kafka Connector `task` array for resource %s: %s", d.Id(), err)
			}
		}
	}

	if !found {
		return diag.Errorf("cannot read Kafka Connector resource with Id: %s not found in a Kafka Connectors list", d.Id())
	}

	return nil
}

func resourceKafkaConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	connectorName := d.Get("connector_name").(string)

	config := make(aiven.KafkaConnectorConfig)
	for k, cS := range d.Get("config").(map[string]interface{}) {
		config[k] = cS.(string)
	}

	err := m.(*aiven.Client).KafkaConnectors.Create(project, serviceName, config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, connectorName))

	return resourceKafkaConnectorRead(ctx, d, m)
}

func resourceKafkaConnectorDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, service, name, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = m.(*aiven.Client).KafkaConnectors.Delete(project, service, name)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, connectorName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	config := make(aiven.KafkaConnectorConfig)
	for k, cS := range d.Get("config").(map[string]interface{}) {
		config[k] = cS.(string)
	}

	_, err = m.(*aiven.Client).KafkaConnectors.Update(project, serviceName, connectorName, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaConnectorRead(ctx, d, m)
}
