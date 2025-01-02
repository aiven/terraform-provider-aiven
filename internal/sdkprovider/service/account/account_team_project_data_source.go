package account

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAccountTeamProjectRead),
		Description: "The Account Team Project data source provides information about the existing Account Team Project.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamProjectSchema,
			"account_id", "team_id", "project_name"),
	}
}

func datasourceAccountTeamProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)

	d.SetId(schemautil.BuildResourceID(accountID, teamID, projectName))

	return resourceAccountTeamProjectRead(ctx, d, client)
}
