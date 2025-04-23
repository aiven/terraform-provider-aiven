package account

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/accountteam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const deprecationMessage = `
!> **Teams have been deprecated and are being migrated to groups**
**On 30 September 2024** the Account Owners team will transition to super admin. Super admin have full access to the organization.
The Account Owners and super admin are synced, so the removal of the Account Owners team will have no impact on existing permissions.
**From 4 November 2024** you won't be able to create new teams or update existing ones. Existing teams will be migrated to groups after
this date. **On 2 December 2024** all teams will be deleted and the teams feature will be completely removed. [View the
migration guide](https://aiven.io/docs/tools/terraform/howto/migrate-from-teams-to-groups) for more information on the changes and migrating to groups.

~> **Important**
You can't delete the Account Owners team. **Deleting all other teams in your organization will disable the teams feature.**
You won't be able to create new teams or access your Account Owners team.
`

var aivenAccountTeamSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The unique account id",
	},
	"team_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The auto-generated unique account team id",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The account team name",
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation",
	},
	"update_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of last update",
	},
}

func ResourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		Description:   `Creates and manages a team.`,
		CreateContext: common.WithGenClient(resourceAccountTeamCreate),
		ReadContext:   common.WithGenClient(resourceAccountTeamRead),
		UpdateContext: common.WithGenClient(resourceAccountTeamUpdate),
		DeleteContext: common.WithGenClient(resourceAccountTeamDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenAccountTeamSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func resourceAccountTeamCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		name      = d.Get("name").(string)
		accountID = d.Get("account_id").(string)
	)

	resp, err := client.AccountTeamCreate(ctx, accountID, &accountteam.AccountTeamCreateIn{
		TeamName: name,
	})
	if err != nil {
		return err
	}

	if resp.AccountId == nil {
		return fmt.Errorf("account team create response missing account_id field")
	}

	d.SetId(schemautil.BuildResourceID(*resp.AccountId, resp.TeamId))

	return resourceAccountTeamRead(ctx, d, client)
}

func resourceAccountTeamRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.AccountTeamGet(ctx, accountID, teamID)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	if err = schemautil.ResourceDataSet(
		d, resp, aivenAccountTeamSchema,
		schemautil.RenameAlias("team_name", "name"),
		schemautil.SetForceNew("account_id", accountID),
	); err != nil {
		return err
	}

	return nil
}

func resourceAccountTeamUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.AccountTeamUpdate(ctx, accountID, teamID, &accountteam.AccountTeamUpdateIn{
		TeamName: d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	if resp.AccountId == nil {
		return fmt.Errorf("account team update response missing account_id field")
	}

	d.SetId(schemautil.BuildResourceID(*resp.AccountId, resp.TeamId))

	return resourceAccountTeamRead(ctx, d, client)
}

func resourceAccountTeamDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	err = client.AccountTeamDelete(ctx, accountID, teamID)
	if common.IsCritical(err) {
		return err
	}

	return nil
}
