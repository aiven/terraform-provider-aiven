// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceProjectUser() *schema.Resource {
	return &schema.Resource{
		Read: datasourceProjectUserRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenProjectUserSchema,
			"project", "email"),
	}
}

func datasourceProjectUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	email := d.Get("email").(string)

	users, invitations, err := client.ProjectUsers.List(projectName)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Email == email {
			d.SetId(buildResourceID(projectName, email))
			return resourceProjectUserRead(d, m)
		}
	}

	for _, invitation := range invitations {
		if invitation.UserEmail == email {
			d.SetId(buildResourceID(projectName, email))
			return resourceProjectUserRead(d, m)
		}
	}

	return fmt.Errorf("project user %s/%s not found", projectName, email)
}
