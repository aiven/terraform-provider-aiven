// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceProjectUser() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceProjectUserRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenProjectUserSchema, "project", "email"),
	}
}

func datasourceProjectUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	email := d.Get("email").(string)

	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, email))
	d.Set("project", projectName)
	d.Set("email", email)
	if user != nil {
		d.Set("member_type", user.MemberType)
		d.Set("accepted", true)
	} else {
		d.Set("member_type", invitation.MemberType)
		d.Set("accepted", false)
	}
	return nil
}
