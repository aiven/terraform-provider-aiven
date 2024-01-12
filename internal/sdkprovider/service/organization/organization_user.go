package organization

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
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
		Deprecated:  "This field is deprecated and will be removed in the next major release. ",
	},
	"accepted": {
		Type:     schema.TypeBool,
		Computed: true,
		Description: "This is a boolean flag that determines whether an invitation was accepted or not by the user. " +
			"`false` value means that the invitation was sent to the user but not yet accepted. `true` means that" +
			" the user accepted the invitation and now a member of an organization.",
		Deprecated: "This field is deprecated and will be removed in the next major release. ",
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
		Description: `
The Organization User resource allows the creation and management of an Aiven Organization User.

During the creation of ` + "`aiven_organization_user`" + `resource, an email invitation will be sent
to a user using ` + "`user_email`" + ` address. If the user accepts an invitation, they will become
a member of the organization. The deletion of ` + "`aiven_organization_user`" + ` will not only
delete the invitation if one was sent but not yet accepted by the user, it will also 
eliminate the member from the organization if one has accepted an invitation previously.
`,
		CreateContext: resourceOrganizationUserCreate,
		ReadContext:   resourceOrganizationUserRead,
		DeleteContext: resourceOrganizationUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOrganizationUserSchema,
		DeprecationMessage: `
This resource is deprecated, please use aiven_organization_user datasource instead. 
Invitation of organization users is not supported anymore via Terraform. Therefore 
creation of this resource is not supported anymore. We reccoemnd to use WebUI to create
and organization user invitation. And upon receiving an invitation, a user can accept it 
using WebUI. Once accepted, the user will become a member of the organization and will 
be able to access it via Terraform.
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
func resourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	organizationID, userEmail, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	rm, err := client.OrganizationUser.List(ctx, organizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, user := range rm.Users {
		userInfo := user.UserInfo

		if userInfo.UserEmail == userEmail {
			if err := d.Set("organization_id", organizationID); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("user_email", userInfo.UserEmail); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("create_time", user.JoinTime.String()); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("user_id", user.UserID); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func resourceOrganizationUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	organizationID, userEmail, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	found := true

	// delete organization user invitation
	err = client.OrganizationUserInvitations.Delete(ctx, organizationID, userEmail)
	if err != nil {
		if !aiven.IsNotFound(err) {
			return diag.FromErr(err)
		}

		found = false
	}

	r, err := client.OrganizationUser.List(ctx, organizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(r.Users) == 0 {
		return nil
	}

	// delete organization user
	for _, u := range r.Users {
		userInfo := u.UserInfo

		if userInfo.UserEmail == userEmail {
			err = client.OrganizationUser.Delete(ctx, organizationID, u.UserID)
			if common.IsCritical(err) {
				return diag.FromErr(err)
			}
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("user with email %q is not a part of the organization %q", userEmail, organizationID)
	}

	return nil
}
