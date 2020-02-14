package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

var aivenAccountTeamSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Required:    true,
	},
	"team_id": {
		Type:        schema.TypeString,
		Description: "Account team id",
		Computed:    true,
	},
	"name": {
		Type:        schema.TypeString,
		Description: "Account team name",
		Required:    true,
	},
	"create_time": {
		Type:        schema.TypeString,
		Description: "Time of creation",
		Optional:    true,
		Computed:    true,
	},
	"update_time": {
		Type:        schema.TypeString,
		Description: "Time of last update",
		Optional:    true,
		Computed:    true,
	},
}

func resourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountTeamCreate,
		Read:   resourceAccountTeamRead,
		Update: resourceAccountTeamUpdate,
		Delete: resourceAccountTeamDelete,
		Exists: resourceAccountTeamExists,
		Importer: &schema.ResourceImporter{
			State: resourceAccountTeamState,
		},

		Schema: aivenAccountTeamSchema,
	}
}

func resourceAccountTeamCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	r, err := client.AccountTeams.Create(
		accountId,
		aiven.AccountTeam{
			Name: name,
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(r.Team.AccountId, r.Team.Id))

	return resourceAccountTeamRead(d, m)
}

func resourceAccountTeamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId := splitResourceID2(d.Id())
	r, err := client.AccountTeams.Get(accountId, teamId)
	if err != nil {
		return err
	}

	if err := d.Set("account_id", r.Team.AccountId); err != nil {
		return err
	}
	if err := d.Set("team_id", r.Team.Id); err != nil {
		return err
	}
	if err := d.Set("name", r.Team.Name); err != nil {
		return err
	}
	if err := d.Set("create_time", r.Team.CreateTime.String()); err != nil {
		return err
	}
	if err := d.Set("update_time", r.Team.UpdateTime.String()); err != nil {
		return err
	}

	return nil
}

func resourceAccountTeamUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	accountId, teamId := splitResourceID2(d.Id())

	r, err := client.AccountTeams.Update(accountId, teamId, aiven.AccountTeam{
		Name: d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(r.Team.AccountId, r.Team.Id))

	return resourceAccountTeamRead(d, m)
}

func resourceAccountTeamDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId := splitResourceID2(d.Id())

	err := client.AccountTeams.Delete(accountId, teamId)
	if err != nil {
		return err
	}

	return nil
}

func resourceAccountTeamExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	_, err := client.AccountTeams.Get(splitResourceID2(d.Id()))
	if err != nil {
		return false, err
	}

	return resourceExists(err)
}

func resourceAccountTeamState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceAccountTeamRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
