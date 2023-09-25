package account

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountAuthenticationRead,
		Description: "The Account Authentication data source provides information about the existing Aiven Account Authentication.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountAuthenticationSchema,
			"account_id", "name"),
	}
}

func datasourceAccountAuthenticationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	r, err := client.AccountAuthentications.List(ctx, accountID)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, a := range r.AuthenticationMethods {
		if a.AuthenticationMethodName == name {
			d.SetId(schemautil.BuildResourceID(a.AccountID, a.AuthenticationMethodID))
			return resourceAccountAuthenticationRead(ctx, d, m)
		}
	}

	return diag.Errorf("account authentication %s not found", name)
}
