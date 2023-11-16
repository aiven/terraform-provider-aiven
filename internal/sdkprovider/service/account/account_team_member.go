package account

import (
	"context"
	"log"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAccountTeamMemberSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The unique account id").ForceNew().Build(),
	},
	"team_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("An account team id").ForceNew().Build(),
	},
	"user_email": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: userconfig.Desc(
			"Is a user email address that first will be invited, and after accepting an invitation, he " +
				"or she becomes a member of a team. Should be lowercase.",
		).ForceNew().Build(),
		ValidateFunc: schemautil.ValidateEmailAddress,
	},
	"invited_by_user_email": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The email address that invited this user.",
	},
	"accepted": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "is a boolean flag that determines whether an invitation was accepted or not by the user. `false` value means that the invitation was sent to the user but not yet accepted. `true` means that the user accepted the invitation and now a member of an account team.",
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation",
	},
}

func ResourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		Description: `
The Account Team Member resource allows the creation and management of an Aiven Account Team Member.

During the creation of ` + "`aiven_account_team_member`" + `resource, an email invitation will be sent
to a user using ` + "`user_email`" + ` address. If the user accepts an invitation, he or she will become
a member of the account team. The deletion of ` + "`aiven_account_team_member`" + ` will not only
delete the invitation if one was sent but not yet accepted by the user, it will also 
eliminate an account team member if one has accepted an invitation previously.
`,
		CreateContext: resourceAccountTeamMemberCreate,
		ReadContext:   resourceAccountTeamMemberRead,
		DeleteContext: resourceAccountTeamMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenAccountTeamMemberSchema,
		DeprecationMessage: "This resource is deprecated",
	}
}

func resourceAccountTeamMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	accountID := d.Get("account_id").(string)
	teamID := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	err := client.AccountTeamMembers.Invite(
		ctx,
		accountID,
		teamID,
		userEmail,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, m)
}

func resourceAccountTeamMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var found bool
	client := m.(*aiven.Client)

	accountID, teamID, userEmail, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.AccountTeamInvites.List(ctx, accountID, teamID)
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
		rm, err := client.AccountTeamMembers.List(ctx, accountID, teamID)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, member := range rm.Members {
			if member.UserEmail == userEmail {
				found = true

				if err := d.Set("account_id", accountID); err != nil {
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
		log.Printf("[WARNING] cannot find user invitation for %s", d.Id())
		if !d.Get("accepted").(bool) {
			log.Printf("[DEBUG] resending account team member invitation ")
			return resourceAccountTeamMemberCreate(ctx, d, m)
		}
	}

	return nil
}

func resourceAccountTeamMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, teamID, userEmail, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// delete account team user invitation
	err = client.AccountTeamInvites.Delete(ctx, accountID, teamID, userEmail)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	r, err := client.AccountTeamMembers.List(ctx, accountID, teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(r.Members) == 0 {
		return nil
	}

	// delete account team member
	found := false
	for _, m := range r.Members {
		if m.UserEmail == userEmail {
			err = client.AccountTeamMembers.Delete(ctx, accountID, teamID, m.UserId)
			if err != nil && !aiven.IsNotFound(err) {
				return diag.FromErr(err)
			}
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("user with email %q is not a part of the team %q", userEmail, teamID)
	}

	return nil
}
