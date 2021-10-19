// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceServiceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceUserRead,
		Description: "The Service User data source provides information about the existing Aiven Service User.",
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceServiceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range list {
		if u.Username == userName {
			d.SetId(buildResourceID(projectName, serviceName, userName))
			return resourceServiceUserRead(ctx, d, m)
		}
	}

	return diag.Errorf("service user %s/%s/%s not found",
		projectName, serviceName, userName)
}
