// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenServiceIntegrationEndpointSchema = map[string]*schema.Schema{
	"project": {
		Description: "Project the service integration endpoint belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_name": {
		ForceNew:    true,
		Description: "Name of the service integration endpoint",
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_type": {
		Description: "Type of the service integration endpoint",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_config": {
		Description: "Integration endpoint specific backend configuration",
		Computed:    true,
		Type:        schema.TypeMap,
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"datadog_user_config": {
		Description: "Datadog specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["datadog"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"prometheus_user_config": {
		Description: "Prometheus specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["prometheus"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"rsyslog_user_config": {
		Description: "rsyslog specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["rsyslog"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_elasticsearch_logs_user_config": {
		Description: "external elasticsearch specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_elasticsearch_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_aws_cloudwatch_logs_user_config": {
		Description: "external AWS CloudWatch Logs specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_aws_cloudwatch_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_google_cloud_logging_user_config": {
		Description: "external Google Cloud Logginig specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_google_cloud_logging"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_kafka_user_config": {
		Description: "external Kafka specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_kafka"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"jolokia_user_config": {
		Description: "Jolokia specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["jolokia"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"signalfx_user_config": {
		Description: "Signalfx specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["signalfx"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_schema_registry_user_config": {
		Description: "External schema registry specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_schema_registry"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_aws_cloudwatch_metrics_user_config": {
		Description: "External AWS cloudwatch mertrics specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("endpoint")["external_aws_cloudwatch_metrics"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
}

func resourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service Integration Endpoint resource allows the creation and management of Aiven Service Integration Endpoints.",
		CreateContext: resourceServiceIntegrationEndpointCreate,
		ReadContext:   resourceServiceIntegrationEndpointRead,
		UpdateContext: resourceServiceIntegrationEndpointUpdate,
		DeleteContext: resourceServiceIntegrationEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceIntegrationEndpointState,
		},

		Schema: aivenServiceIntegrationEndpointSchema,
	}
}

func resourceServiceIntegrationEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	endpointType := d.Get("endpoint_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("endpoint", endpointType, true, d)
	endpoint, err := client.ServiceIntegrationEndpoints.Create(
		projectName,
		aiven.CreateServiceIntegrationEndpointRequest{
			EndpointName: d.Get("endpoint_name").(string),
			EndpointType: endpointType,
			UserConfig:   userConfig,
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(projectName, endpoint.EndpointID))

	return resourceServiceIntegrationEndpointRead(ctx, d, m)
}

func resourceServiceIntegrationEndpointRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	endpoint, err := client.ServiceIntegrationEndpoints.Get(projectName, endpointID)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	err = copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceIntegrationEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	endpointType := d.Get("endpoint_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("endpoint", endpointType, false, d)
	_, err := client.ServiceIntegrationEndpoints.Update(
		projectName,
		endpointID,
		aiven.UpdateServiceIntegrationEndpointRequest{
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceIntegrationEndpointRead(ctx, d, m)
}

func resourceServiceIntegrationEndpointDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	err := client.ServiceIntegrationEndpoints.Delete(projectName, endpointID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceIntegrationEndpointState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<endpoint_id>", d.Id())
	}

	projectName, endpointID := splitResourceID2(d.Id())
	endpoint, err := client.ServiceIntegrationEndpoints.Get(projectName, endpointID)
	if err != nil {
		return nil, err
	}

	err = copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	endpoint *aiven.ServiceIntegrationEndpoint,
	project string,
) error {
	d.Set("project", project)
	d.Set("endpoint_name", endpoint.EndpointName)
	endpointType := endpoint.EndpointType
	d.Set("endpoint_type", endpointType)
	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat("endpoint", endpointType, endpoint.UserConfig)
	if len(userConfig) > 0 {
		d.Set(endpointType+"_user_config", userConfig)
	}
	// Must coerse all values into strings
	endpointConfig := map[string]string{}
	if len(endpoint.EndpointConfig) > 0 {
		for key, value := range endpoint.EndpointConfig {
			endpointConfig[key] = fmt.Sprintf("%v", value)
		}
	}
	d.Set("endpoint_config", endpointConfig)

	return nil
}
