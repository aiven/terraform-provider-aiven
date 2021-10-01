// Package aiven Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAzurePrivatelink() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzurePrivatelinkRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenAzurePrivatelinkSchema, "project", "service_name"),
	}
}

func datasourceAzurePrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(buildResourceID(projectName, serviceName))

	return resourceAzurePrivatelinkRead(ctx, d, m)
}
