// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
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

	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return err
	}

	return copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
}
