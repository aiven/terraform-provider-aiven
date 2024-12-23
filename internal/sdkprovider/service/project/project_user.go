package project

import (
	"context"
	"errors"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenProjectUserSchema = map[string]*schema.Schema{
	"project": schemautil.CommonSchemaProjectReference,
	"email": {
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		Description:  userconfig.Desc("Email address of the user in lowercase.").ForceNew().Build(),
		ValidateFunc: schemautil.ValidateEmailAddress,
	},
	"member_type": {
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("Project membership type.").PossibleValuesString(account.MemberTypeChoices()...).Build(),
	},
	"accepted": {
		Computed:    true,
		Type:        schema.TypeBool,
		Description: "Whether the user has accepted the request to join the project. Users get an invite and become project members after accepting the invite.",
	},
}

func ResourceProjectUser() *schema.Resource {
	return &schema.Resource{
		Description: `Creates and manages an Aiven project member.

**This resource is deprecated.** Use ` + "`aiven_organization_permission`" + ` and
[migrate existing aiven_project_user resources](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/update-deprecated-resources) 
to the new resource.
		`,
		CreateContext: resourceProjectUserCreate,
		ReadContext:   resourceProjectUserRead,
		UpdateContext: resourceProjectUserUpdate,
		DeleteContext: resourceProjectUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenProjectUserSchema,
		DeprecationMessage: "This resource is deprecated. Use aiven_organization_permission instead.",
	}
}

// isProjectUserAlreadyInvited return true if user already been invited to the project
func isProjectUserAlreadyInvited(err error) bool {
	var e aiven.Error
	if errors.As(err, &e) {
		if strings.Contains(e.Message, "already been invited to this project") && e.Status == 409 {
			return true
		}
	}
	return false
}

func resourceProjectUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	email := d.Get("email").(string)
	err := client.ProjectUsers.Invite(
		ctx,
		projectName,
		aiven.CreateProjectInvitationRequest{
			UserEmail:  email,
			MemberType: d.Get("member_type").(string),
		},
	)
	if err != nil && !isProjectUserAlreadyInvited(err) {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, email))
	if err := d.Set("accepted", false); err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectUserRead(ctx, d, m)
}

func resourceProjectUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, invitation, err := client.ProjectUsers.Get(ctx, projectName, email)
	if err != nil {
		if aiven.IsNotFound(err) && !d.Get("accepted").(bool) {
			return resourceProjectUserCreate(ctx, d, m)
		}
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", email); err != nil {
		return diag.FromErr(err)
	}
	if user != nil {
		if err := d.Set("member_type", user.MemberType); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("accepted", true); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("member_type", invitation.MemberType); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("accepted", false); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceProjectUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	memberType := d.Get("member_type").(string)
	err = client.ProjectUsers.UpdateUserOrInvitation(
		ctx,
		projectName,
		email,
		aiven.UpdateProjectUserOrInvitationRequest{
			MemberType: memberType,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectUserRead(ctx, d, m)
}

func resourceProjectUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, invitation, err := client.ProjectUsers.Get(ctx, projectName, email)
	if err != nil {
		return diag.FromErr(err)
	}

	// delete user if exists
	if user != nil {
		err := client.ProjectUsers.DeleteUser(ctx, projectName, email)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 ||
				!strings.Contains(e.Message, "User does not exist") ||
				!strings.Contains(e.Message, "User not found") {

				return diag.FromErr(err)
			}
		}
	}

	// delete invitation if exists
	if invitation != nil {
		err := client.ProjectUsers.DeleteInvitation(ctx, projectName, email)
		if common.IsCritical(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
