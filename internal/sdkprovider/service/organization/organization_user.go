package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationUserSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The unique organization ID").ForceNew().Build(),
	},
	"user_email": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: userconfig.Desc("This is a user email address that first will be invited, " +
			"and after accepting an invitation, they become a member of the organization. Should be lowercase.",
		).ForceNew().Build(),
		ValidateFunc: schemautil.ValidateEmailAddress,
	},
	"invited_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The email address of the user who sent an invitation to the user.",
		Deprecated:  "This field is deprecated and will be removed in the next major release.",
	},
	"accepted": {
		Type:     schema.TypeBool,
		Computed: true,
		Description: "This is a boolean flag that determines whether an invitation was accepted or not by the user. " +
			"`false` value means that the invitation was sent to the user but not yet accepted. `true` means that" +
			" the user accepted the invitation and now a member of an organization.",
		Deprecated: "This field is deprecated and will be removed in the next major release.",
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation",
	},
	"user_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The unique organization user ID",
	},
}

func ResourceOrganizationUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven Organization user.",
		CreateContext: resourceOrganizationUserCreate,
		ReadContext:   common.WithGenClient(resourceOrganizationUserRead),
		DeleteContext: common.WithGenClient(resourceOrganizationUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOrganizationUserSchema,
		DeprecationMessage: `
Users cannot be invited to an organization using Terraform.
Use the Aiven Console to [invite users to your organization](https://aiven.io/docs/platform/howto/manage-org-users).
After the user accepts the invite you can get their information using the aiven_organization_user data source. You can manage
user access to projects with the aiven_organization_user_group, aiven_organization_user_group_member,
and aiven_organization_permission resources.
		`,
	}
}

// resourceOrganizationUserCreate create is not supported anymore
func resourceOrganizationUserCreate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("creation of organization user is not supported anymore via Terraform. " +
		"Please use WebUI to create an organization user invitation. And upon receiving an invitation, " +
		"a user can accept it using WebUI. Once accepted, the user will become a member of the organization " +
		"and will be able to access it via Terraform using datasource `aiven_organization_user`")
}

// resourceOrganizationUserRead reads the properties of an Aiven Organization User and provides them to Terraform
func resourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, userEmail, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.OrganizationUserList(ctx, organizationID)
	if err != nil {
		return err
	}

	for _, user := range resp {
		if user.UserInfo.UserEmail == userEmail {
			if err = d.Set("organization_id", organizationID); err != nil {
				return err
			}
			if err = d.Set("user_email", userEmail); err != nil {
				return err
			}
			if err = d.Set("create_time", user.JoinTime.String()); err != nil {
				return err
			}
			if err = d.Set("user_id", user.UserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceOrganizationUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	invitationFound := true

	organizationID, userEmail, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	// delete organization user invitation
	err = client.OrganizationUserInvitationDelete(ctx, organizationID, userEmail)
	if err != nil {
		if !avngen.IsNotFound(err) {
			return err
		}

		invitationFound = false
	}

	resp, err := client.OrganizationUserList(ctx, organizationID)
	if err != nil {
		return err
	}

	if len(resp) == 0 {
		return nil
	}

	// delete organization user
	for _, u := range resp {
		userInfo := u.UserInfo

		if userInfo.UserEmail == userEmail {
			err = client.OrganizationUserDelete(ctx, organizationID, u.UserId)
			if err != nil && !avngen.IsNotFound(err) {
				return err
			}

			return nil
		}
	}

	if invitationFound {
		return nil
	}

	return fmt.Errorf("user with email %q is not a part of the organization %q", userEmail, organizationID)
}
