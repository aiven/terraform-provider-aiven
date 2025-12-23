package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenKafkaConnectorSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"connector_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The Kafka connector name.").ForceNew().Build(),
	},
	"config": {
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "The Kafka connector configuration parameters.",
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
		Description: "The version of the Kafka connector.",
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
					Description: "The task ID of the task.",
					Computed:    true,
				},
			},
		},
	},
}

func ResourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages Aiven for Apache Kafka® [connectors](https://aiven.io/docs/products/kafka/kafka-connect/concepts/list-of-connector-plugins).
Source connectors let you import data from an external system into a Kafka topic. Sink connectors let you export data from a topic to an external system.

You can use connectors with any Aiven for Apache Kafka® service that is integrated with an Aiven for Apache Kafka® Connect service.
`,
		CreateContext: common.WithGenClientDiag(resourceKafkaConnectorCreate),
		ReadContext:   common.WithGenClientDiag(resourceKafkaConnectorRead),
		UpdateContext: common.WithGenClientDiag(resourceKafkaTConnectorUpdate),
		DeleteContext: common.WithGenClientDiag(resourceKafkaConnectorDelete),
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
func customizeDiffKafkaConnectorConfigName() func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
	return func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
		connectorName := diff.Get("connector_name").(string)
		config := make(map[string]string)
		for k, cS := range diff.Get("config").(map[string]any) {
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
	return func(_ context.Context, _, newValue, _ interface{}) bool {
		for k, cS := range newValue.(map[string]interface{}) {
			if k == "name" && cS.(string) != "" {
				return true
			}
		}
		return false
	}
}

func flattenKafkaConnectorTasks(tasks []kafkaconnect.TaskOut) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tasks))

	for i, taskS := range tasks {
		task := map[string]interface{}{
			"connector": taskS.Connector,
			"task":      taskS.Task,
		}

		result[i] = task
	}

	return result
}

func resourceKafkaConnectorRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, serviceName, connectorName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	connectors, err := client.ServiceKafkaConnectList(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	var found bool
	for _, r := range connectors {
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
			if err := d.Set("plugin_doc_url", r.Plugin.DocUrl); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_doc_url` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_title", r.Plugin.Title); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_title` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_type", string(r.Plugin.Type)); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_type` for resource %s: %s", d.Id(), err)
			}
			if err := d.Set("plugin_version", r.Plugin.Version); err != nil {
				return diag.Errorf("error setting Kafka Connector `plugin_version` for resource %s: %s", d.Id(), err)
			}

			tasks := flattenKafkaConnectorTasks(r.Tasks)
			if err := d.Set("task", tasks); err != nil {
				return diag.Errorf("error setting Kafka Connector `task` array for resource %s: %s", d.Id(), err)
			}
		}
	}

	if !found {
		d.SetId("")
		return nil
	}

	return nil
}

const pollCreateTimeout = time.Minute * 2

func resourceKafkaConnectorCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	connectorName := d.Get("connector_name").(string)

	configRaw := d.Get("config").(map[string]any)
	config := make(map[string]string, len(configRaw))
	for k, v := range configRaw {
		config[k] = v.(string)
	}

	// Sometimes this method returns 404: Not Found for the given serviceName
	// Since the client has its own retries for various scenarios
	// we retry here 404 only
	err := retry.RetryContext(ctx, pollCreateTimeout, func() *retry.RetryError {
		_, err := client.ServiceKafkaConnectCreateConnector(ctx, project, serviceName, &config)
		if err == nil {
			return nil
		}

		var e avngen.Error
		if errors.As(err, &e) && e.Status == 201 {
			// 201: Created
			return nil
		}

		return &retry.RetryError{
			Err:       err,
			Retryable: avngen.IsNotFound(err), // retries 404 only
		}
	})
	if err != nil {
		log.Printf("[DEBUG] Kafka Connectors create err %s", err.Error())
		return diag.FromErr(fmt.Errorf("kafka connector create error: %w", err))
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, connectorName))

	// Wait for connector to appear in the list after creation
	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{"IN_PROGRESS"},
		Target:  []string{"OK"},
		Refresh: func() (any, string, error) {
			connectors, err := client.ServiceKafkaConnectList(ctx, project, serviceName)
			if err != nil {
				log.Printf("[DEBUG] Kafka Connectors list waiter err %s", err.Error())
				// connector was already created successfully, any error here is transient
				return nil, "IN_PROGRESS", nil
			}

			for _, r := range connectors {
				if r.Name == connectorName {
					return &r, "OK", nil
				}
			}

			return nil, "IN_PROGRESS", nil
		},
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}

	_, err = stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Kafka Connector to be created: %s", err)
	}

	return resourceKafkaConnectorRead(ctx, d, client)
}

func resourceKafkaConnectorDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, service, name, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceKafkaConnectDeleteConnector(ctx, project, service, name)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTConnectorUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, serviceName, connectorName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	configRaw := d.Get("config").(map[string]any)
	config := make(map[string]string, len(configRaw))
	for k, v := range configRaw {
		config[k] = v.(string)
	}

	_, err = client.ServiceKafkaConnectEditConnector(ctx, project, serviceName, connectorName, &config)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaConnectorRead(ctx, d, client)
}
