package account

import (
	"context"
	"log"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/accountteammember"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
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
Adds a user as a team member.

During the creation of this resource, an invite is sent to the address specified in ` + "`user_email`" + `.
The user is added to the team after they accept the invite. Deleting ` + "`aiven_account_team_member`" + `
deletes the pending invite if not accepted or removes the user from the team if they already accepted the invite.
`,
		CreateContext: common.WithGenClient(resourceAccountTeamMemberCreate),
		ReadContext:   common.WithGenClient(resourceAccountTeamMemberRead),
		DeleteContext: common.WithGenClient(resourceAccountTeamMemberDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenAccountTeamMemberSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func resourceAccountTeamMemberCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		accountID = d.Get("account_id").(string)
		teamID    = d.Get("team_id").(string)
		userEmail = d.Get("user_email").(string)
	)

	if err := client.AccountTeamMembersInvite(
		ctx,
		accountID,
		teamID,
		&accountteammember.AccountTeamMembersInviteIn{Email: userEmail},
	); err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(accountID, teamID, userEmail))

	return resourceAccountTeamMemberRead(ctx, d, client)
}

func resourceAccountTeamMemberRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, userEmail, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.AccountTeamMembersList(ctx, accountID, teamID)
	if err != nil {
		return err
	}

	for _, invite := range resp {
		if invite.UserEmail == userEmail {
			if err = schemautil.ResourceDataSet(
				d, invite, aivenAccountTeamMemberSchema,
				schemautil.SetForceNew("account_id", accountID),
			); err != nil {
				return err
			}

			// if a user is in the invitations list, it means invitation was sent but not yet accepted
			if err = d.Set("accepted", false); err != nil {
				return err
			}

			return nil
		}
	}

	respTI, err := client.AccountTeamMembersList(ctx, accountID, teamID)
	if err != nil {
		return err
	}

	for _, member := range respTI {
		if member.UserEmail == userEmail {
			if err = schemautil.ResourceDataSet(
				d, member, aivenAccountTeamMemberSchema,
				schemautil.SetForceNew("account_id", accountID),
			); err != nil {
				return err
			}

			// when a user accepts an invitation, it will appear in the member's list
			// and disappear from invitations list
			if err = d.Set("accepted", true); err != nil {
				return err
			}

			return nil
		}
	}

	log.Printf("[WARNING] cannot find user invitation for %s", d.Id())
	if !d.Get("accepted").(bool) {
		log.Printf("[DEBUG] resending account team member invitation ")
		return resourceAccountTeamMemberCreate(ctx, d, client)
	}

	return nil
}

func resourceAccountTeamMemberDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, userEmail, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	// delete account team user invitation
	if err = client.AccountTeamMemberCancelInvite(ctx, accountID, teamID, userEmail); common.IsCritical(err) {
		return err
	}

	resp, err := client.AccountTeamMembersList(ctx, accountID, teamID)
	if err != nil {
		return err
	}

	for _, m := range resp {
		if m.UserEmail == userEmail {
			if err = client.AccountTeamMembersDelete(ctx, accountID, teamID, m.UserId); common.IsCritical(err) {
				return err
			}
		}
	}

	// we don't need to return an error if a user is not found in the members list
	// because it means that a user was invited but not yet accepted an invitation,
	// and we already deleted an invitation above

	return nil
}
