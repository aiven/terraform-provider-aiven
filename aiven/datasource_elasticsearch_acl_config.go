// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceElasticsearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLConfigRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenElasticsearchACLConfigSchema, "project", "service_name"),
	}
}

func datasourceElasticsearchACLConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if acl != nil {
		d.SetId(buildResourceID(projectName, serviceName))

		return resourceElasticsearchACLConfigRead(ctx, d, m)
	}

	return diag.Errorf("acl config %s/%s not found", projectName, serviceName)
}
