package opensearch

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpensearchACLRuleRead,
		Description: "The Opensearch ACL Rule data source provides information about an existing Aiven Opensearch ACL Rule.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenOpensearchACLRuleSchema, "project", "service_name", "username", "index", "permission"),
	}
}

func datasourceOpensearchACLRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)

	r, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, found := resourceElasticsearchACLRuleGetPermissionFromACLResponse(r.ElasticSearchACLConfig, username, index); !found {
		return diag.Errorf("acl rule %s/%s/%s/%s not found", projectName, serviceName, username, index)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username, index))

	return resourceOpensearchACLRuleRead(ctx, d, m)
}
