// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceServiceUser() *schema.Resource {
	return &schema.Resource{
		Read: datasourceServiceUserRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceServiceUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(projectName, serviceName)
	if err != nil {
		return err
	}

	for _, u := range list {
		if u.Username == userName {
			d.SetId(buildResourceID(projectName, serviceName, userName))
			return resourceServiceUserRead(d, m)
		}
	}

	return fmt.Errorf("service user %s/%s/%s not found",
		projectName, serviceName, userName)
}
