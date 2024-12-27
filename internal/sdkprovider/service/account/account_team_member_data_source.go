package account

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAccountTeamMemberRead),
		Description: "The Account Team Member data source provides information about the existing Aiven Account Team Member.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamMemberSchema,
			"account_id", "team_id", "user_email"),
	}
}

func datasourceAccountTeamMemberRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	d.SetId(schemautil.BuildResourceID(accountID, teamID, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, client)
}
