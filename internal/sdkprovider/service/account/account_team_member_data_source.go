package account

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAccountTeamMemberRead,
		Description: "The Account Team Member data source provides information about the existing Aiven Account Team Member.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamMemberSchema,
			"account_id", "team_id", "user_email"),
	}
}

func datasourceAccountTeamMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	d.SetId(schemautil.BuildResourceID(accountID, teamID, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, m)
}
