package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpensearchACLConfigRead,
		Description: "The Opensearch ACL Config data source provides information about an existing " +
			"Aiven Opensearch ACL Config.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOpensearchACLConfigSchema, "project", "service_name",
		),
	}
}

func datasourceOpensearchACLConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if acl != nil {
		d.SetId(schemautil.BuildResourceID(projectName, serviceName))

		return resourceOpensearchACLConfigRead(ctx, d, m)
	}

	return diag.Errorf("acl config %s/%s not found", projectName, serviceName)
}
