package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOrganizationUserRead,
		Description: "The Organization User data source provides information about the existing Aiven" +
			" Organization User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOrganizationUserSchema, "organization_id", "user_email",
		),
	}
}

func datasourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organizationID := d.Get("organization_id").(string)
	userEmail := d.Get("user_email").(string)

	d.SetId(schemautil.BuildResourceID(organizationID, userEmail))

	return resourceOrganizationUserRead(ctx, d, m)
}
