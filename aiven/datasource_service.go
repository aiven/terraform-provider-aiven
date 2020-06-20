// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceService() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceSchema, "project", "service_name"),
	}
}

func datasourceServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(buildResourceID(projectName, serviceName))

	services, err := client.Services.List(projectName)
	for _, service := range services {
		if service.Name == serviceName {
			return resourceServiceRead(d, m)
		}
	}

	if err != nil {
		return err
	}

	return fmt.Errorf("service %s/%s not found", projectName, serviceName)
}
