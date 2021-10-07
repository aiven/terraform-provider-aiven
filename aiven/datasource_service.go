// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceService() *schema.Resource {
	return &schema.Resource{
		ReadContext:        datasourceServiceRead,
		Description:        "The Service datasource provides information about specific Aiven Services.",
		DeprecationMessage: "Please use the specific service datasources instead of this datasource.",
		Schema:             resourceSchemaAsDatasourceSchema(aivenServiceSchema, "project", "service_name"),
	}
}

func datasourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(buildResourceID(projectName, serviceName))

	services, err := client.Services.List(projectName)
	for _, service := range services {
		if service.Name == serviceName {
			return resourceServiceRead(ctx, d, m)
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Errorf("service %s/%s not found", projectName, serviceName)
}
