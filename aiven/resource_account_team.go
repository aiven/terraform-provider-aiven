package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
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
	"invite_users": {
		Type:        schema.TypeString,
		Description: "A comma separated list of users who should be invited to a team",
		Optional:    true,
	},
	"member": {
		Type:        schema.TypeSet,
		Description: "List of team members",
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"user_id": {
					Type:        schema.TypeString,
					Description: "Team member user Id",
					Optional:    true,
					Computed:    true,
				},
				"real_name": {
					Type:        schema.TypeString,
					Description: "Team member real name",
					Optional:    true,
					Computed:    true,
				},
				"user_email": {
					Type:        schema.TypeString,
					Description: "Team member user email",
					Optional:    true,
					Computed:    true,
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
			},
		},
	},
	"invite": {
		Type:        schema.TypeSet,
		Description: "List of team members",
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"user_email": {
					Type:        schema.TypeString,
					Description: "Team invite user email",
					Optional:    true,
					Computed:    true,
				},
				"invited_by_user_email": {
					Type:        schema.TypeString,
					Description: "Team invited by user email",
					Optional:    true,
					Computed:    true,
				},
				"status": {
					Type:        schema.TypeString,
					Description: "Team invitation status",
					Optional:    true,
					Computed:    true,
				},
				"create_time": {
					Type:        schema.TypeString,
					Description: "Time of creation",
					Optional:    true,
					Computed:    true,
				},
			},
		},
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

	for _, email := range strings.Split(d.Get("invite_users").(string), ",") {
		err := client.AccountTeamMembers.Invite(
			accountId,
			r.Team.Id,
			email)
		if err != nil {
			return err
		}
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

	mr, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return err
	}

	if err := d.Set("member", flattenAccountTeamMembers(mr)); err != nil {
		return fmt.Errorf("cannot set account team members(%s), err: %s", d.Id(), err)
	}

	mi, err := client.AccountTeamInvites.List(accountId, teamId)
	if err != nil {
		return err
	}

	if err := d.Set("invite", flattenAccountTeamInvites(mi)); err != nil {
		return fmt.Errorf("cannot set account team invites (%s), err: %s", d.Id(), err)
	}

	return nil
}

func flattenAccountTeamMembers(r *aiven.AccountTeamMembersResponse) []map[string]interface{} {
	var members []map[string]interface{}

	for _, memberS := range r.Members {
		member := map[string]interface{}{
			"user_email":  memberS.UserEmail,
			"real_name":   memberS.RealName,
			"user_id":     memberS.UserId,
			"create_time": memberS.CreateTime.String(),
			"update_time": memberS.UpdateTime.String(),
		}

		members = append(members, member)
	}

	return members
}

func flattenAccountTeamInvites(r *aiven.AccountTeamInvitesResponse) []map[string]interface{} {
	var invites []map[string]interface{}

	for _, inviteS := range r.Invites {
		invite := map[string]interface{}{
			"user_email":            inviteS.UserEmail,
			"invited_by_user_email": inviteS.InvitedByUserEmail,
			"create_time":           inviteS.CreateTime.String(),
			"status":                "sent",
		}

		invites = append(invites, invite)
	}

	return invites
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

	members, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return err
	}

	invites, err := client.AccountTeamInvites.List(accountId, teamId)
	if err != nil {
		return err
	}

	for _, email := range strings.Split(d.Get("invite_users").(string), ",") {
		if !isUserInMembersOrInvites(email, invites, members) {
			err := client.AccountTeamMembers.Invite(
				accountId,
				r.Team.Id,
				email)
			if err != nil {
				return err
			}
		}
	}

	d.SetId(buildResourceID(r.Team.AccountId, r.Team.Id))

	return resourceAccountTeamRead(d, m)
}

func isUserInMembersOrInvites(
	email string,
	i *aiven.AccountTeamInvitesResponse,
	m *aiven.AccountTeamMembersResponse) bool {
	for _, invite := range i.Invites {
		if invite.UserEmail == email {
			return true
		}
	}

	for _, member := range m.Members {
		if member.UserEmail == email {
			return true
		}
	}

	return false
}

func resourceAccountTeamDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId := splitResourceID2(d.Id())

	listMembers, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return err
	}

	for _, member := range listMembers.Members {
		err := client.AccountTeamMembers.Delete(accountId, teamId, member.UserId)
		if err != nil {
			return err
		}
	}

	err = client.AccountTeams.Delete(accountId, teamId)
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
