// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceIntegrationEndpointCreate,
		Read:   resourceServiceIntegrationEndpointRead,
		Update: resourceServiceIntegrationEndpointUpdate,
		Delete: resourceServiceIntegrationEndpointDelete,
		Exists: resourceServiceIntegrationEndpointExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceIntegrationEndpointState,
		},

		Schema: map[string]*schema.Schema{
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
			"datadog_user_config": {
				Description: "Datadog specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["endpoint"]["datadog"].(map[string]interface{})),
				},
				MaxItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
		},
	}
}

func resourceServiceIntegrationEndpointCreate(d *schema.ResourceData, m interface{}) error {
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
		return err
	}

	d.SetId(buildResourceID(projectName, endpoint.EndpointID))
	return copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
}

func resourceServiceIntegrationEndpointRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	endpoint, err := client.ServiceIntegrationEndpoints.Get(projectName, endpointID)
	if err != nil {
		return err
	}

	return copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
}

func resourceServiceIntegrationEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	endpointType := d.Get("endpoint_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("endpoint", endpointType, false, d)
	endpoint, err := client.ServiceIntegrationEndpoints.Update(
		projectName,
		endpointID,
		aiven.UpdateServiceIntegrationEndpointRequest{
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return err
	}

	return copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
}

func resourceServiceIntegrationEndpointDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	return client.ServiceIntegrationEndpoints.Delete(projectName, endpointID)
}

func resourceServiceIntegrationEndpointExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, endpointID := splitResourceID2(d.Id())
	_, err := client.ServiceIntegrationEndpoints.Get(projectName, endpointID)
	return resourceExists(err)
}

func resourceServiceIntegrationEndpointState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<endpoint_id>", d.Id())
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
	if userConfig != nil && len(userConfig) > 0 {
		d.Set(endpointType+"_user_config", userConfig)
	}

	return nil
}
