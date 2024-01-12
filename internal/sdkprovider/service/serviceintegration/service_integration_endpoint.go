package serviceintegration

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/integration_endpoint"
)

var integrationEndpointTypes = []string{
	"datadog",
	"prometheus",
	"rsyslog",
	"external_elasticsearch_logs",
	"external_opensearch_logs",
	"external_aws_cloudwatch_logs",
	"external_google_cloud_logging",
	"external_kafka",
	"jolokia",
	"external_schema_registry",
	"external_aws_cloudwatch_metrics",
	"external_google_cloud_bigquery",
	"external_postgresql",
}

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
		Description: "Type of the service integration endpoint. Possible values: " +
			schemautil.JoinQuoted(integrationEndpointTypes, ", ", "`"),
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice(integrationEndpointTypes, false),
	},
	"endpoint_config": {
		Description: "Integration endpoint specific backend configuration",
		Computed:    true,
		Type:        schema.TypeMap,
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"datadog_user_config":                         integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeDatadog),
	"prometheus_user_config":                      integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypePrometheus),
	"rsyslog_user_config":                         integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeRsyslog),
	"external_elasticsearch_logs_user_config":     integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalElasticsearchLogs),
	"external_opensearch_logs_user_config":        integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalOpensearchLogs),
	"external_aws_cloudwatch_logs_user_config":    integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalAwsCloudwatchLogs),
	"external_google_cloud_logging_user_config":   integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalGoogleCloudLogging),
	"external_kafka_user_config":                  integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalKafka),
	"jolokia_user_config":                         integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeJolokia),
	"external_schema_registry_user_config":        integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalSchemaRegistry),
	"external_aws_cloudwatch_metrics_user_config": integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalAwsCloudwatchMetrics),
	"external_google_cloud_bigquery":              integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalGoogleCloudBigquery),
	"external_postgresql":                         integration_endpoint.GetUserConfig(schemautil.ServiceIntegrationEndpointTypeExternalPostgresql),
}

func ResourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service Integration Endpoint resource allows the creation and management of Aiven Service Integration Endpoints.",
		CreateContext: resourceServiceIntegrationEndpointCreate,
		ReadContext:   resourceServiceIntegrationEndpointRead,
		UpdateContext: resourceServiceIntegrationEndpointUpdate,
		DeleteContext: resourceServiceIntegrationEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         aivenServiceIntegrationEndpointSchema,
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.ServiceIntegrationEndpoint(),
	}
}

func resourceServiceIntegrationEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	endpointType := d.Get("endpoint_type").(string)

	userConfig, err := schemautil.ExpandServiceIntegrationEndpoint(endpointType, d)
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint, err := client.ServiceIntegrationEndpoints.Create(
		ctx,
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

	d.SetId(schemautil.BuildResourceID(projectName, endpoint.EndpointID))

	return resourceServiceIntegrationEndpointRead(ctx, d, m)
}

func resourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint, err := client.ServiceIntegrationEndpoints.Get(ctx, projectName, endpointID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceIntegrationEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpointType := d.Get("endpoint_type").(string)

	userConfig, err := schemautil.ExpandServiceIntegrationEndpoint(endpointType, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceIntegrationEndpoints.Update(
		ctx,
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

func resourceServiceIntegrationEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceIntegrationEndpoints.Delete(ctx, projectName, endpointID)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	endpoint *aiven.ServiceIntegrationEndpoint,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("endpoint_name", endpoint.EndpointName); err != nil {
		return err
	}
	endpointType := endpoint.EndpointType
	if err := d.Set("endpoint_type", endpointType); err != nil {
		return err
	}

	userConfig, err := schemautil.FlattenServiceIntegrationEndpoint(endpointType, d, endpoint.UserConfig)
	if err != nil {
		return err
	}

	if len(userConfig) > 0 {
		if err := d.Set(endpointType+"_user_config", userConfig); err != nil {
			return err
		}
	}
	// Must coerse all values into strings
	endpointConfig := map[string]string{}
	if len(endpoint.EndpointConfig) > 0 {
		for key, value := range endpoint.EndpointConfig {
			endpointConfig[key] = fmt.Sprintf("%v", value)
		}
	}

	return d.Set("endpoint_config", endpointConfig)
}
