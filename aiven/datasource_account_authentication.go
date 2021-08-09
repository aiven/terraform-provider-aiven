package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountAuthenticationRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountAuthenticationSchema,
			"account_id", "name"),
	}
}

func datasourceAccountAuthenticationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	r, err := client.AccountAuthentications.List(accountId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, a := range r.AuthenticationMethods {
		if a.Name == name {
			d.SetId(buildResourceID(a.AccountId, a.Id))
			return resourceAccountAuthenticationRead(ctx, d, m)
		}
	}

	return diag.Errorf("account authentication %s not found", name)
}
