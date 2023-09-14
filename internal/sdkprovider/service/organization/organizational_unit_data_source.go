package organization

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOrganizationalUnitRead,
		Description: "The Organizational Unit data source provides information about the existing Aiven " +
			"Organizational Unit.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenOrganizationalUnitSchema, "name"),
	}
}

func datasourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)

	r, err := client.Accounts.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, ac := range r.Accounts {
		if ac.Name == name {
			d.SetId(ac.Id)
			return resourceOrganizationalUnitRead(ctx, d, m)
		}
	}

	return diag.Errorf("organizational unit %s not found", name)
}
