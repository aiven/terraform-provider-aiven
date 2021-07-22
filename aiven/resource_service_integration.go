// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
	"strings"
)

const serviceIntegrationEndpointRegExp = "^[a-zA-Z0-9_-]*\\/{1}[a-zA-Z0-9_-]*$"

var aivenServiceIntegrationSchema = map[string]*schema.Schema{
	"destination_endpoint_id": {
		Description: "Destination endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"destination_service_name": {
		Description: "Destination service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
	"integration_type": {
		Description: "Type of the service integration",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"logs_user_config": {
		Description: "Log integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"mirrormaker_user_config": {
		Description: "Mirrormaker 1 integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["mirrormaker"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_mirrormaker_user_config": {
		Description: "Mirrormaker 2 integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_mirrormaker"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_connect_user_config": {
		Description: "Kafka Connect specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_connect"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_google_cloud_logging_user_config": {
		Description: "External Google Cloud Logging specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["external_google_cloud_logging"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_elasticsearch_logs_user_config": {
		Description: "External Elasticsearch logs specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["external_elasticsearch_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_aws_cloudwatch_logs_user_config": {
		Description: "External AWS Cloudwatch logs specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["external_aws_cloudwatch_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"read_replica_user_config": {
		Description: "PG Read replica specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["read_replica"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"rsyslog_user_config": {
		Description: "RSyslog specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["rsyslog"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"signalfx_user_config": {
		Description: "Signalfx specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["signalfx"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"dashboard_user_config": {
		Description: "Dashboard specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["dashboard"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"datadog_user_config": {
		Description: "Dashboard specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["datadog"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_logs_user_config": {
		Description: "Kafka Logs specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"m3aggregator_user_config": {
		Description: "M3 aggregator specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["m3aggregator"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"m3coordinator_user_config": {
		Description: "M3 coordinator specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["m3coordinator"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"prometheus_user_config": {
		Description: "Prometheus coordinator specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["prometheus"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"metrics_user_config": {
		Description: "Metrics specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["metrics"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"schema_registry_proxy_user_config": {
		Description: "Schema registry proxy specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["schema_registry_proxy"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"external_aws_cloudwatch_metrics_user_config": {
		Description: "External AWS cloudwatch metrics specific user configurable settings",
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["external_aws_cloudwatch_metrics"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"project": {
		Description: "Project the integration belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"source_endpoint_id": {
		Description: "Source endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"source_service_name": {
		Description: "Source service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
}

func resourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceIntegrationCreate,
		ReadContext:   resourceServiceIntegrationRead,
		UpdateContext: resourceServiceIntegrationUpdate,
		DeleteContext: resourceServiceIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceIntegrationState,
		},

		Schema: aivenServiceIntegrationSchema,
	}
}

func plainEndpointID(fullEndpointID *string) *string {
	if fullEndpointID == nil {
		return nil
	}
	_, endpointID := splitResourceID2(*fullEndpointID)
	return &endpointID
}

func resourceServiceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var integration *aiven.ServiceIntegration
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("integration", integrationType, true, d)

	// Some service integrations can be created alongside the service creation, like `read_replica`,
	// for example. And for such cases, we check if a service integration already exists before
	// creating a new one.
	list, err := client.ServiceIntegrations.List(projectName, sourceServiceName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("cannot get a list of service integration: %s", err)
	}

	for _, i := range list {
		if i.SourceService == nil || i.DestinationService == nil || i.ServiceIntegrationID == "" {
			continue
		}

		if i.IntegrationType == integrationType &&
			*i.SourceService == sourceServiceName &&
			*i.DestinationService == destinationServiceName {
			integration = i
			break
		}
	}

	// When service integration does not exist, create a new one
	if integration == nil {
		integration, err = client.ServiceIntegrations.Create(
			projectName,
			aiven.CreateServiceIntegrationRequest{
				DestinationEndpointID: plainEndpointID(optionalStringPointer(d, "destination_endpoint_id")),
				DestinationService:    optionalStringPointer(d, "destination_service_name"),
				IntegrationType:       integrationType,
				SourceEndpointID:      plainEndpointID(optionalStringPointer(d, "source_endpoint_id")),
				SourceService:         optionalStringPointer(d, "source_service_name"),
				UserConfig:            userConfig,
			},
		)
		if err != nil {
			return diag.Errorf("cannot create service integration: %s", err)
		}
	}

	d.SetId(buildResourceID(projectName, integration.ServiceIntegrationID))

	// Checking if there are any differences in user config and if there are
	// differences update service integration
	if userConfig != nil &&
		!cmp.Equal(userConfig, integration.UserConfig, cmpopts.SortMaps(func(a, b string) bool {
			return strings.Compare(a, b) == -1
		})) {
		return resourceServiceIntegrationUpdate(ctx, d, m)

	}

	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		return diag.Errorf("cannot get service integration: %s; id: %s", err, integrationID)
	}

	err = copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
	if err != nil {
		return diag.Errorf("cannot populate TF fields: %s", err)
	}

	return nil
}

func resourceServiceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	integrationType := d.Get("integration_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("integration", integrationType, false, d)

	_, err := client.ServiceIntegrations.Update(
		projectName,
		integrationID,
		aiven.UpdateServiceIntegrationRequest{
			UserConfig: nil,
		},
	)
	if err != nil {
		return diag.Errorf("unable to update service integration: %s; user config: %+v", err, userConfig)
	}

	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	err := client.ServiceIntegrations.Delete(projectName, integrationID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("cannot delete service integration: %s", err)
	}

	return nil
}

func resourceServiceIntegrationState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<integration_id>", d.Id())
	}

	projectName, integrationID := splitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		return nil, err
	}

	err = copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func copyServiceIntegrationPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	integration *aiven.ServiceIntegration,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}

	if integration.DestinationEndpointID != nil {
		if err := d.Set("destination_endpoint_id", buildResourceID(project, *integration.DestinationEndpointID)); err != nil {
			return err
		}
	} else if integration.DestinationService != nil {
		if err := d.Set("destination_service_name", *integration.DestinationService); err != nil {
			return err
		}
	}
	if integration.SourceEndpointID != nil {
		if err := d.Set("source_endpoint_id", buildResourceID(project, *integration.SourceEndpointID)); err != nil {
			return err
		}
	} else if integration.SourceService != nil {
		if err := d.Set("source_service_name", *integration.SourceService); err != nil {
			return err
		}
	}
	integrationType := integration.IntegrationType
	if err := d.Set("integration_type", integrationType); err != nil {
		return err
	}

	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat(
		"integration",
		integrationType,
		integration.UserConfig,
	)
	if len(userConfig) > 0 {
		d.Set(integrationType+"_user_config", userConfig)
	}

	return nil
}
