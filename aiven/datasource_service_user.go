// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceServiceUser() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceUserRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceUserSchema, "project", "service_name", "username"),
	}
}

func datasourceServiceUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	user, err := client.ServiceUsers.Get(projectName, serviceName, userName)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName, userName))
	return copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
}
