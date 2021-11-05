// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamProjectRead,
		Description: "The Account Team Project data source provides information about the existing Account Team Project.",
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamProjectSchema,
			"account_id", "team_id", "project_name"),
	}
}

func datasourceAccountTeamProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)

	d.SetId(buildResourceID(accountId, teamId, projectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}
