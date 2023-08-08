package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAWSPrivatelink() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAWSPrivatelinkRead,
		Description: "The AWS Privatelink resource allows the creation and management of Aiven AWS Privatelink for a services.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenAWSPrivatelinkSchema, "project", "service_name"),
	}
}

func datasourceAWSPrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	return resourceAWSPrivatelinkRead(ctx, d, m)
}
