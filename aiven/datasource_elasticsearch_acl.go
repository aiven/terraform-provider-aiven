// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceElasticsearchACL() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "ElasticsearchACL is deprecated please use ElasticsearchACLConfig and ElasticsearchACLRule",
		Description:        "The Elasticsearch ACL data source provides information about the existing Aiven Elasticsearch ACL for Elasticsearch service.",
		ReadContext:        datasourceElasticsearchACLRead,
		Schema:             resourceSchemaAsDatasourceSchema(aivenElasticsearchACLSchema, "project", "service_name"),
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
		d.SetId(schemautil.BuildResourceID(projectName, serviceName))

		return resourceElasticsearchACLRead(ctx, d, m)
	}

	return diag.Errorf("elasticsearch acl %s/%s not found", projectName, serviceName)
}
