// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceElasticsearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLConfigRead,
		Description: "The Elasticsearch ACL Config data source provides information about an existing Aiven Elasticsearch ACL Config.",
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
		d.SetId(schemautil.BuildResourceID(projectName, serviceName))

		return resourceElasticsearchACLConfigRead(ctx, d, m)
	}

	return diag.Errorf("acl config %s/%s not found", projectName, serviceName)
}
