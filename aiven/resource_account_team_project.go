package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
		Create: resourceAccountTeamProjectCreate,
		Read:   resourceAccountTeamProjectRead,
		Update: resourceAccountTeamProjectUpdate,
		Delete: resourceAccountTeamProjectDelete,
		Exists: resourceAccountTeamProjectExists,
		Importer: &schema.ResourceImporter{
			State: resourceAccountTeamProjectState,
		},

		Schema: aivenAccountTeamProjectSchema,
	}
}

func resourceAccountTeamProjectCreate(d *schema.ResourceData, m interface{}) error {
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
		return err
	}

	d.SetId(buildResourceID(accountId, teamId, projectName))

	return resourceAccountTeamProjectRead(d, m)
}

func resourceAccountTeamProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId, projectName := splitResourceID3(d.Id())
	r, err := client.AccountTeamProjects.List(accountId, teamId)
	if err != nil {
		return err
	}

	var project aiven.AccountTeamProject
	for _, p := range r.Projects {
		if p.ProjectName == projectName {
			project = p
		}
	}

	if project.ProjectName == "" {
		return fmt.Errorf("account team project %s not found", d.Id())
	}

	if err := d.Set("account_id", accountId); err != nil {
		return err
	}
	if err := d.Set("team_id", teamId); err != nil {
		return err
	}
	if err := d.Set("project_name", project.ProjectName); err != nil {
		return err
	}
	if err := d.Set("team_type", project.TeamType); err != nil {
		return err
	}

	return nil
}

func resourceAccountTeamProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId, _ := splitResourceID3(d.Id())
	newProjectName := d.Get("project_name").(string)
	teamType := d.Get("team_type").(string)

	err := client.AccountTeamProjects.Update(accountId, teamId, aiven.AccountTeamProject{
		TeamType:    teamType,
		ProjectName: newProjectName,
	})
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(accountId, teamId, newProjectName))

	return resourceAccountTeamRead(d, m)
}

func resourceAccountTeamProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.AccountTeamProjects.Delete(
		splitResourceID3(d.Id()))
}

func resourceAccountTeamProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	accountId, teamId, projectName := splitResourceID3(d.Id())
	r, err := client.AccountTeamProjects.List(accountId, teamId)
	if err != nil {
		return resourceExists(err)
	}

	for _, p := range r.Projects {
		if p.ProjectName == projectName {
			return true, nil
		}
	}

	return false, nil
}

func resourceAccountTeamProjectState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceAccountTeamProjectRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
