// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenKafkaConnectorSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"connector_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: complex("The kafka connector name.").forceNew().build(),
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

func resourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka connectors resource allows the creation and management of Aiven Kafka connectors.",
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
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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
