package account

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		ValidateFunc: validation.StringInSlice([]string{"admin", "developer", "operator", "read_only"}, false),
		Description:  userconfig.Desc("The Account team project type").PossibleValues("admin", "developer", "operator", "read_only").Build(),
	},
}

func ResourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		Description: `
Links an existing project to an existing team. Both the project and team should have the same ` + "`account_id`" + `.
`,
		CreateContext: resourceAccountTeamProjectCreate,
		ReadContext:   resourceAccountTeamProjectRead,
		UpdateContext: resourceAccountTeamProjectUpdate,
		DeleteContext: resourceAccountTeamProjectDelete,
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

func resourceAccountTeamProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err := client.AccountTeamProjects.Create(
		ctx,
		accountID,
		teamID,
		aiven.AccountTeamProject{
			ProjectName: projectName,
			TeamType:    teamType,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, projectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, teamID, projectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.AccountTeamProjects.List(ctx, accountID, teamID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	var project aiven.AccountTeamProject
	for _, p := range r.Projects {
		if p.ProjectName == projectName {
			project = p
		}
	}

	if project.ProjectName == "" {
		return diag.Errorf("account team project %s not found", d.Id())
	}

	if err := d.Set("account_id", accountID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team_id", teamID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_name", project.ProjectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team_type", project.TeamType); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAccountTeamProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, teamID, _, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	newProjectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err = client.AccountTeamProjects.Update(ctx, accountID, teamID, aiven.AccountTeamProject{
		TeamType:    teamType,
		ProjectName: newProjectName,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, newProjectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, teamID, projectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AccountTeamProjects.Delete(ctx, accountID, teamID, projectName)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}
