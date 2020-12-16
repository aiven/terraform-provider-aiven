package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		CreateContext: resourceAccountTeamMemberCreate,
		ReadContext:   resourceAccountTeamMemberRead,
		UpdateContext: resourceAccountTeamMemberCreate,
		DeleteContext: resourceAccountTeamMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountTeamMemberState,
		},

		Schema: aivenAccountTeamMemberSchema,
	}
}

func resourceAccountTeamMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	err := client.AccountTeamMembers.Invite(
		accountId,
		teamId,
		userEmail)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(accountId, teamId, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, m)
}

func resourceAccountTeamMemberRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var found bool
	client := m.(*aiven.Client)

	accountId, teamId, userEmail := splitResourceID3(d.Id())

	r, err := client.AccountTeamInvites.List(accountId, teamId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, invite := range r.Invites {
		if invite.UserEmail == userEmail {
			found = true

			if err := d.Set("account_id", invite.AccountId); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("team_id", invite.TeamId); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("user_email", invite.UserEmail); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("invited_by_user_email", invite.InvitedByUserEmail); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("create_time", invite.CreateTime.String()); err != nil {
				return diag.FromErr(err)
			}

			// if a user is in the invitations list, it means invitation was sent but not yet accepted
			if err := d.Set("accepted", false); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if !found {
		rm, err := client.AccountTeamMembers.List(accountId, teamId)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, member := range rm.Members {
			if member.UserEmail == userEmail {
				found = true

				if err := d.Set("account_id", accountId); err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("team_id", member.TeamId); err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("user_email", member.UserEmail); err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("create_time", member.CreateTime.String()); err != nil {
					return diag.FromErr(err)
				}

				// when a user accepts an invitation, it will appear in the member's list
				// and disappear from invitations list
				if err := d.Set("accepted", true); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	if !found {
		return diag.Errorf("cannot find user invitation for %s", d.Id())
	}

	return nil
}

func resourceAccountTeamMemberDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, teamId, userEmail := splitResourceID3(d.Id())

	// delete account team user invitation
	err := client.AccountTeamInvites.Delete(accountId, teamId, userEmail)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	r, err := client.AccountTeamMembers.List(accountId, teamId)
	if err != nil {
		return diag.FromErr(err)
	}

	// delete account team member
	for _, m := range r.Members {
		if m.UserEmail == userEmail {
			err = client.AccountTeamMembers.Delete(splitResourceID3(d.Id()))
			if err != nil && !aiven.IsNotFound(err) {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func resourceAccountTeamMemberState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAccountTeamMemberRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get account team member: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
