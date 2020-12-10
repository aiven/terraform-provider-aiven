package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

var aivenKafkaConnectorSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the kafka connector to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the kafka connector to",
		Required:    true,
		ForceNew:    true,
	},
	"connector_name": {
		Type:        schema.TypeString,
		Description: "Kafka connector name",
		Required:    true,
		ForceNew:    true,
	},
	"config": {
		Type:        schema.TypeMap,
		Description: "Kafka Connector configuration parameters",
		Required:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"plugin_author": {
		Type:        schema.TypeString,
		Description: "Kafka connector author",
		Computed:    true,
	},
	"plugin_class": {
		Type:        schema.TypeString,
		Description: "Kafka connector Java class",
		Computed:    true,
	},
	"plugin_doc_url": {
		Type:        schema.TypeString,
		Description: "Kafka connector documentation URL",
		Computed:    true,
	},
	"plugin_title": {
		Type:        schema.TypeString,
		Description: "Kafka connector title",
		Computed:    true,
	},
	"plugin_type": {
		Type:        schema.TypeString,
		Description: "Kafka connector type",
		Computed:    true,
	},
	"plugin_version": {
		Type:        schema.TypeString,
		Description: "Kafka connector version",
		Computed:    true,
	},
	"task": {
		Type:        schema.TypeSet,
		Description: "List of tasks of a connector",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connector": {
					Type:        schema.TypeString,
					Description: "Related connector name",
					Computed:    true,
				},
				"task": {
					Type:        schema.TypeInt,
					Description: "Task id / number",
					Computed:    true,
				},
			},
		},
	},
}

func resourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKafkaConnectorCreate,
		ReadContext:   resourceKafkaConnectorRead,
		UpdateContext: resourceKafkaTConnectorUpdate,
		DeleteContext: resourceKafkaConnectorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaConnectorState,
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenKafkaConnectorSchema,
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

func resourceKafkaConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, connectorName := splitResourceID3(d.Id())

	w := &KafkaConnectorWaiter{
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
	}

	timeout := d.Timeout(schema.TimeoutRead)
	res, err := w.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("cannot read Kafka Connector List resource %s: %s", d.Id(), err)
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
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(project, serviceName, connectorName))

	return resourceKafkaConnectorRead(ctx, d, m)
}

func resourceKafkaConnectorDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := m.(*aiven.Client).KafkaConnectors.Delete(splitResourceID3(d.Id()))
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, connectorName := splitResourceID3(d.Id())

	config := make(aiven.KafkaConnectorConfig)
	for k, cS := range d.Get("config").(map[string]interface{}) {
		config[k] = cS.(string)
	}

	_, err := m.(*aiven.Client).KafkaConnectors.Update(project, serviceName, connectorName, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaConnectorRead(ctx, d, m)
}

func resourceKafkaConnectorState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceKafkaConnectorRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka connector: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

// KafkaConnectorWaiter is used to wait for Aiven Kafka Connector list to be available
type KafkaConnectorWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaConnectorWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		list, err := w.Client.KafkaConnectors.List(w.Project, w.ServiceName)
		if err != nil {
			log.Printf("[DEBUG] Kafka Connectors list waiter err %s", err.Error())

			if aiven.IsNotFound(err) {
				return nil, "", err
			}

			return nil, "IN_PROGRESS", nil
		}

		return list, "OK", nil
	}
}

// Conf sets up the configuration to refresh.
func (w *KafkaConnectorWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Kafka Connector waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending: []string{
			"IN_PROGRESS",
		},
		Target: []string{
			"OK",
		},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}
}
