package account

import (
	"context"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
=======

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamRead,
		Description: "The Account Team data source provides information about the existing Account Team.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamSchema,
			"account_id", "name"),
	}
}

func datasourceAccountTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	r, err := client.AccountTeams.List(accountID)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, t := range r.Teams {
		if t.Name == name {
			d.SetId(schemautil.BuildResourceID(t.AccountId, t.Id))
			return resourceAccountTeamRead(ctx, d, m)
		}
	}

	return diag.Errorf("account team %s not found", name)
}
