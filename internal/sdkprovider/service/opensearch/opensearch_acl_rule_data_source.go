package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOpenSearchACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpenSearchACLRuleRead,
		Description: "The OpenSearch ACL Rule data source provides information about an existing Aiven OpenSearch ACL Rule.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenOpenSearchACLRuleSchema, "project", "service_name", "username", "index", "permission"),
	}
}

func datasourceOpenSearchACLRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)

	r, err := client.OpenSearchACLs.Get(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, found := resourceOpenSearchACLRuleGetPermissionFromACLResponse(r.OpenSearchACLConfig, username, index); !found {
		return diag.Errorf("acl rule %s/%s/%s/%s not found", projectName, serviceName, username, index)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username, index))

	return resourceOpenSearchACLRuleRead(ctx, d, m)
}
