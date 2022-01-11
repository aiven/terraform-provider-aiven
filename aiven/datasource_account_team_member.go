// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamMemberRead,
		Description: "The Account Team Member data source provides information about the existing Aiven Account Team Member.",
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamMemberSchema,
			"account_id", "team_id", "user_email"),
	}
}

func datasourceAccountTeamMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	d.SetId(schemautil.BuildResourceID(accountId, teamId, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, m)
}
