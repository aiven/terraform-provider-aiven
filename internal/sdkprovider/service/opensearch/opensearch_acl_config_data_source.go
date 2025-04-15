package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOpenSearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpenSearchACLConfigRead,
		Description: "Gets information about access control for an Aiven for OpenSearch® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenOpenSearchACLConfigSchema, "project", "service_name"),
	}
}

func datasourceOpenSearchACLConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.OpenSearchACLs.Get(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if acl != nil {
		d.SetId(schemautil.BuildResourceID(projectName, serviceName))

		return resourceOpenSearchACLConfigRead(ctx, d, m)
	}

	return diag.Errorf("acl config %s/%s not found", projectName, serviceName)
}
