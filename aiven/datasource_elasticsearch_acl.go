// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceElasticsearchACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenElasticsearchACLSchema,
			"project", "service_name"),
	}
}

func datasourceElasticsearchACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if acl != nil {
		d.SetId(buildResourceID(projectName, serviceName))

		return resourceElasticsearchACLRead(ctx, d, m)
	}

	return diag.Errorf("elasticsearch acl %s/%s not found",
		projectName, serviceName)
}
