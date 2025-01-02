package account

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/aiven/go-client-codegen/handler/accountteam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAccountTeamProjectSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The unique account id",
	},
	"team_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "An account team id",
	},
	"project_name": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The name of an already existing project",
	},
	"team_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(accountteam.TeamTypeChoices(), false),
		Description:  userconfig.Desc("The Account team project type").PossibleValuesString(account.TeamTypeChoices()...).Build(),
	},
}

func ResourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		Description: `
Links an existing project to an existing team. Both the project and team should have the same ` + "`account_id`" + `.
`,
		CreateContext: common.WithGenClient(resourceAccountTeamProjectCreate),
		ReadContext:   common.WithGenClient(resourceAccountTeamProjectRead),
		UpdateContext: common.WithGenClient(resourceAccountTeamProjectUpdate),
		DeleteContext: common.WithGenClient(resourceAccountTeamProjectDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAccountTeamProjectSchema,
		DeprecationMessage: `
This resource is deprecated. Use aiven_organization_user_group instead.

You can't delete the Account Owners team. Deleting all other teams in your organization will disable the teams feature. You won't be able to create new teams or access your Account Owners team.

On 2 December 2024 all teams will be deleted and the teams feature will be completely removed. View the
migration guide for more information: https://aiven.io/docs/tools/terraform/howto/migrate-from-teams-to-groups.
`,
	}
}

func resourceAccountTeamProjectCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	if err := client.AccountTeamProjectAssociate(
		ctx,
		accountID,
		teamID,
		projectName,
		&accountteam.AccountTeamProjectAssociateIn{
			TeamType: accountteam.TeamType(teamType),
		},
	); err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, projectName))

	return resourceAccountTeamProjectRead(ctx, d, client)
}

func resourceAccountTeamProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, projectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.AccountTeamProjectList(ctx, accountID, teamID)
	if err != nil {
		return err
	}

	var project accountteam.ProjectOut
	for _, p := range resp {
		if p.ProjectName == projectName {
			project = p
		}
	}

	if project.ProjectName == "" {
		return fmt.Errorf("account team project %q not found", d.Id())
	}

	if err = d.Set("account_id", accountID); err != nil {
		return err
	}
	if err = d.Set("team_id", teamID); err != nil {
		return err
	}
	if err = d.Set("project_name", project.ProjectName); err != nil {
		return err
	}
	if err = d.Set("team_type", project.TeamType); err != nil {
		return err
	}

	return nil
}

func resourceAccountTeamProjectUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, _, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	newProjectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err = client.AccountTeamProjectAssociationUpdate(
		ctx,
		accountID,
		teamID,
		newProjectName,
		&accountteam.AccountTeamProjectAssociationUpdateIn{
			TeamType: accountteam.TeamType(teamType),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, newProjectName))

	return resourceAccountTeamProjectRead(ctx, d, client)
}

func resourceAccountTeamProjectDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, projectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.AccountTeamProjectDisassociate(ctx, accountID, teamID, projectName)
	if common.IsCritical(err) {
		return err
	}

	return nil
}
