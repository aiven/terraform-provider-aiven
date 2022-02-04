// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamRead,
		Description: "The Account Team data source provides information about the existing Account Team.",
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamSchema,
			"account_id", "name"),
	}
}

func datasourceAccountTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	r, err := client.AccountTeams.List(accountId)
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
