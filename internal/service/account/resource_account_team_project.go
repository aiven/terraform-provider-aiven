package account

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Description:  schemautil.Complex("The Account team project type").PossibleValues("admin", "developer", "operator", "read_only").Build(),
	},
}

func ResourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		Description: `
The Account Team Project resource allows the creation and management of an Account Team Project.

It is intended to link an existing project to the existing account team.
It is important to note that the project should have an ` + "`account_id`" + ` property set equal to the
account team you are trying to link to this project.
`,
		CreateContext: resourceAccountTeamProjectCreate,
		ReadContext:   resourceAccountTeamProjectRead,
		UpdateContext: resourceAccountTeamProjectUpdate,
		DeleteContext: resourceAccountTeamProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: aivenAccountTeamProjectSchema,
	}
}

func resourceAccountTeamProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err := client.AccountTeamProjects.Create(
		accountId,
		teamId,
		aiven.AccountTeamProject{
			ProjectName: projectName,
			TeamType:    teamType,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountId, teamId, projectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, teamId, projectName := schemautil.SplitResourceID3(d.Id())
	r, err := client.AccountTeamProjects.List(accountId, teamId)
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

	if err := d.Set("account_id", accountId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team_id", teamId); err != nil {
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

	accountId, teamId, _ := schemautil.SplitResourceID3(d.Id())
	newProjectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err := client.AccountTeamProjects.Update(accountId, teamId, aiven.AccountTeamProject{
		TeamType:    teamType,
		ProjectName: newProjectName,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountId, teamId, newProjectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.AccountTeamProjects.Delete(schemautil.SplitResourceID3(d.Id()))
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}
