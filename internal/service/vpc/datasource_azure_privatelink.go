package vpc

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceAzurePrivatelink() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzurePrivatelinkRead,
		Description: "The Azure Privatelink resource allows the creation and management of " +
			"Aiven Azure Privatelink for a services.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenAzurePrivatelinkSchema, "project", "service_name",
		),
	}
}

func datasourceAzurePrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	return resourceAzurePrivatelinkRead(ctx, d, m)
}
