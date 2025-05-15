package account

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAccountTeamRead),
		Description: "The Account Team data source provides information about the existing Account Team.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountTeamSchema,
			"account_id", "name"),
	}
}

func datasourceAccountTeamRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	resp, err := client.AccountTeamList(ctx, accountID)
	if err != nil {
		return err
	}

	for _, at := range resp {
		if at.TeamName == name {
			if at.AccountId == nil {
				return fmt.Errorf("account team %q not found", name)
			}

			d.SetId(schemautil.BuildResourceID(*at.AccountId, at.TeamId))

			return resourceAccountTeamRead(ctx, d, client)
		}
	}

	return fmt.Errorf("account team %q not found", name)
}
