package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var aivenAccountTeamMemberSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Required:    true,
		ForceNew:    true,
	},
	"team_id": {
		Type:        schema.TypeString,
		Description: "Account team id",
		Required:    true,
		ForceNew:    true,
	},
	"user_email": {
		Type:        schema.TypeString,
		Description: "Team invite user email",
		Required:    true,
		ForceNew:    true,
	},
	"invited_by_user_email": {
		Type:        schema.TypeString,
		Description: "Team invited by user email",
		Optional:    true,
		Computed:    true,
	},
	"accepted": {
		Type:        schema.TypeBool,
		Description: "Team member invitation status",
		Optional:    true,
		Computed:    true,
	},
	"create_time": {
		Type:        schema.TypeString,
		Description: "Time of creation",
		Optional:    true,
		Computed:    true,
	},
}

func resourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountTeamMemberCreate,
		Read:   resourceAccountTeamMemberRead,
		Update: resourceAccountTeamMemberCreate,
		Delete: resourceAccountTeamMemberDelete,
		Exists: resourceAccountTeamMemberExists,
		Importer: &schema.ResourceImporter{
			State: resourceAccountTeamMemberState,
		},

		Schema: aivenAccountTeamMemberSchema,
	}
}

func resourceAccountTeamMemberCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	err := client.AccountTeamMembers.Invite(
		accountId,
		teamId,
		userEmail)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(accountId, teamId, userEmail))

	return resourceAccountTeamMemberRead(d, m)
}

func resourceAccountTeamMemberRead(d *schema.ResourceData, m interface{}) error {
	var found bool
	client := m.(*aiven.Client)

	accountId, teamId, userEmail := splitResourceID3(d.Id())

	r, err := client.AccountTeamInvites.List(accountId, teamId)
	if err != nil {
		return err
	}

	for _, invite := range r.Invites {
		if invite.UserEmail == userEmail {
			found = true

			if err := d.Set("account_id", invite.AccountId); err != nil {
				return err
			}
			if err := d.Set("team_id", invite.TeamId); err != nil {
				return err
			}
			if err := d.Set("user_email", invite.UserEmail); err != nil {
				return err
			}
			if err := d.Set("invited_by_user_email", invite.InvitedByUserEmail); err != nil {
				return err
			}
			if err := d.Set("create_time", invite.CreateTime.String()); err != nil {
				return err
			}

			// if a user is in the invitations list, it means invitation was sent but not yet accepted
			if err := d.Set("accepted", false); err != nil {
				return err
			}
		}
	}

	if !found {
		rm, err := client.AccountTeamMembers.List(accountId, teamId)
		if err != nil {
			return err
		}

		for _, member := range rm.Members {
			if member.UserEmail == userEmail {
				found = true

				if err := d.Set("account_id", accountId); err != nil {
					return err
				}
				if err := d.Set("team_id", member.TeamId); err != nil {
					return err
				}
				if err := d.Set("user_email", member.UserEmail); err != nil {
					return err
				}
				if err := d.Set("create_time", member.CreateTime.String()); err != nil {
					return err
				}

				// when a user accepts an invitation, it will appear in the member's list
				// and disappear from invitations list
				if err := d.Set("accepted", true); err != nil {
					return err
				}
			}
		}
	}

	if !found {
		return fmt.Errorf("cannot find user invitation for %s", d.Id())
	}

	return nil
}

func resourceAccountTeamMemberDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	accountId, teamId, userEmail := splitResourceID3(d.Id())

	// delete account team user invitation
	err := client.AccountTeamInvites.Delete(accountId, teamId, userEmail)
	if err != nil {
		if err.(aiven.Error).Status != 404 {
			return err
		}
	}

	r, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return err
	}

	// delete account team member
	for _, m := range r.Members {
		if m.UserEmail == userEmail {
			err = client.AccountTeamMembers.Delete(splitResourceID3(d.Id()))
			if err != nil {
				if err.(aiven.Error).Status != 404 {
					return err
				}
			}
		}
	}

	return nil
}

func resourceAccountTeamMemberExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	accountId, teamId, userEmail := splitResourceID3(d.Id())

	members, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return false, err
	}

	invites, err := client.AccountTeamInvites.List(accountId, teamId)
	if err != nil {
		return false, err
	}

	if isUserInMembersOrInvites(userEmail, invites, members) {
		return true, nil
	}

	return false, nil
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

func resourceAccountTeamMemberState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceAccountTeamMemberRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
