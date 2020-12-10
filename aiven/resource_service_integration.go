// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

var aivenServiceIntegrationSchema = map[string]*schema.Schema{
	"destination_endpoint_id": {
		Description: "Destination endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
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
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("integration", integrationType, true, d)
	integration, err := client.ServiceIntegrations.Create(
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
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(projectName, integration.ServiceIntegrationID))

	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
	if err != nil {
		return diag.FromErr(err)
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
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	err := client.ServiceIntegrations.Delete(projectName, integrationID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
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
