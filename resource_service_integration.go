// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceIntegrationCreate,
		Read:   resourceServiceIntegrationRead,
		Update: resourceServiceIntegrationUpdate,
		Delete: resourceServiceIntegrationDelete,
		Exists: resourceServiceIntegrationExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceIntegrationState,
		},

		Schema: map[string]*schema.Schema{
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
						GetUserConfigSchema("integration")["logs"].(map[string]interface{})),
				},
				MaxItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
			"mirrormaker_user_config": {
				Description: "Mirrormaker integration specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						GetUserConfigSchema("integration")["mirrormaker"].(map[string]interface{})),
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
		},
	}
}

func plainEndpointID(fullEndpointID *string) *string {
	if fullEndpointID == nil {
		return nil
	}
	_, endpointID := splitResourceID2(*fullEndpointID)
	return &endpointID
}

func resourceServiceIntegrationCreate(d *schema.ResourceData, m interface{}) error {
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
		return err
	}

	d.SetId(buildResourceID(projectName, integration.ServiceIntegrationID))
	return copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
}

func resourceServiceIntegrationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		return err
	}

	return copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
}

func resourceServiceIntegrationUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	integrationType := d.Get("integration_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("integration", integrationType, false, d)
	integration, err := client.ServiceIntegrations.Update(
		projectName,
		integrationID,
		aiven.UpdateServiceIntegrationRequest{
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return err
	}

	return copyServiceIntegrationPropertiesFromAPIResponseToTerraform(d, integration, projectName)
}

func resourceServiceIntegrationDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	return client.ServiceIntegrations.Delete(projectName, integrationID)
}

func resourceServiceIntegrationExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, integrationID := splitResourceID2(d.Id())
	_, err := client.ServiceIntegrations.Get(projectName, integrationID)
	return resourceExists(err)
}

func resourceServiceIntegrationState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<integration_id>", d.Id())
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
	d.Set("project", project)
	if integration.DestinationEndpointID != nil {
		d.Set("destination_endpoint_id", buildResourceID(project, *integration.DestinationEndpointID))
	} else if integration.DestinationService != nil {
		d.Set("destination_service_name", *integration.DestinationService)
	}
	if integration.SourceEndpointID != nil {
		d.Set("source_endpoint_id", buildResourceID(project, *integration.SourceEndpointID))
	} else if integration.SourceService != nil {
		d.Set("source_service_name", *integration.SourceService)
	}
	integrationType := integration.IntegrationType
	d.Set("integration_type", integrationType)
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
