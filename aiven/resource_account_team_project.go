package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenAccountTeamProjectSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Required:    true,
	},
	"team_id": {
		Type:        schema.TypeString,
		Description: "Account team id",
		Required:    true,
	},
	"project_name": {
		Type:        schema.TypeString,
		Description: "Account team project name",
		Optional:    true,
	},
	"team_type": {
		Type:         schema.TypeString,
		Description:  "Account team project type, can one of the following values: admin, developer, operator and read_only",
		Optional:     true,
		ValidateFunc: validation.StringInSlice([]string{"admin", "developer", "operator", "read_only"}, false),
	},
}

func resourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountTeamProjectCreate,
		ReadContext:   resourceAccountTeamProjectRead,
		UpdateContext: resourceAccountTeamProjectUpdate,
		DeleteContext: resourceAccountTeamProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountTeamProjectState,
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

	d.SetId(buildResourceID(accountId, teamId, projectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, teamId, projectName := splitResourceID3(d.Id())
	r, err := client.AccountTeamProjects.List(accountId, teamId)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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

	accountId, teamId, _ := splitResourceID3(d.Id())
	newProjectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err := client.AccountTeamProjects.Update(accountId, teamId, aiven.AccountTeamProject{
		TeamType:    teamType,
		ProjectName: newProjectName,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(accountId, teamId, newProjectName))

	return resourceAccountTeamProjectRead(ctx, d, m)
}

func resourceAccountTeamProjectDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.AccountTeamProjects.Delete(splitResourceID3(d.Id()))
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAccountTeamProjectState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAccountTeamProjectRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get account team project: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
