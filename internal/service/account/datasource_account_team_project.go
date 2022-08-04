package account

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamProjectRead,
		Description: "The Account Team Project data source provides information about the existing " +
			"Account Team Project.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamProjectSchema,
			"account_id", "team_id", "project_name"),
	}
}

func datasourceAccountTeamProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)

	d.SetId(schemautil.BuildResourceID(accountID, teamID, projectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}
