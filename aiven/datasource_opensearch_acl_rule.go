// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpensearchACLRuleRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenOpensearchACLRuleSchema, "project", "service_name", "username", "index", "permission"),
	}
}

func datasourceOpensearchACLRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)

	r, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, found := resourceElasticsearchACLRuleGetPermissionFromACLResponse(r.ElasticSearchACLConfig, username, index); !found {
		return diag.Errorf("opensearch acl rule %s/%s/%s/%s not found", projectName, serviceName, username, index)
	}

	d.SetId(buildResourceID(projectName, serviceName, username, index))

	return resourceElasticsearchACLRuleRead(ctx, d, m)
}
