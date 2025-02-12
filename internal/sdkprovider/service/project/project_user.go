package project

import (
	"context"
	"errors"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/aiven/go-client-codegen/handler/project"
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
		CreateContext: common.WithGenClient(resourceProjectUserCreate),
		ReadContext:   common.WithGenClient(resourceProjectUserRead),
		UpdateContext: common.WithGenClient(resourceProjectUserUpdate),
		DeleteContext: common.WithGenClient(resourceProjectUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenProjectUserSchema,
		DeprecationMessage: "Use aiven_organization_permission instead.",
	}
}

func resourceProjectUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		projectName = d.Get("project").(string)
		email       = d.Get("email").(string)
		memberType  = d.Get("member_type").(string)
	)

	err := client.ProjectInvite(ctx, projectName, &project.ProjectInviteIn{
		MemberType: project.MemberType(memberType),
		UserEmail:  email,
	})

	if err != nil && !isProjectUserAlreadyInvited(err) {
		return err
	}

	d.SetId(schemautil.BuildResourceID(projectName, email))

	if err = d.Set("accepted", false); err != nil {
		return err
	}

	return resourceProjectUserRead(ctx, d, client)
}

func resourceProjectUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	pul, err := client.ProjectUserList(ctx, projectName)
	if err != nil {
		return err
	}

	for _, user := range pul.Users {
		if user.UserEmail == email {
			if err = d.Set("member_type", string(user.MemberType)); err != nil {
				return err
			}

			if err = d.Set("accepted", true); err != nil {
				return err
			}

			return nil
		}
	}

	for _, invitation := range pul.Invitations {
		if invitation.InvitedUserEmail == email {
			if err = d.Set("member_type", string(invitation.MemberType)); err != nil {
				return err
			}

			if err = d.Set("accepted", false); err != nil {
				return err
			}

			return nil
		}
	}

	if !d.Get("accepted").(bool) {
		return resourceProjectUserCreate(ctx, d, client)
	}

	return schemautil.ResourceReadHandleNotFound(errors.New("project user not found"), d)
}

func resourceProjectUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	memberType := d.Get("member_type").(string)
	err = client.ProjectUserUpdate(
		ctx,
		projectName,
		email,
		&project.ProjectUserUpdateIn{MemberType: project.MemberType(memberType)},
	)

	if err == nil {
		return resourceProjectUserRead(ctx, d, client)
	}

	if common.IsCritical(err) {
		return err
	}

	// if user not found, delete the user invite and re-invite
	if err = client.ProjectInviteDelete(ctx, projectName, email); err != nil {
		return err
	}

	if err = client.ProjectInvite(ctx, projectName, &project.ProjectInviteIn{
		MemberType: project.MemberType(memberType),
		UserEmail:  email,
	}); err != nil {
		return err
	}

	return resourceProjectUserRead(ctx, d, client)
}

func resourceProjectUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, email, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	pul, err := client.ProjectUserList(ctx, projectName)
	if err != nil {
		return err
	}

	// delete user if exists
	for _, user := range pul.Users {
		if user.UserEmail == email {
			if err = client.ProjectUserRemove(ctx, projectName, email); err != nil {
				var e avngen.Error
				if errors.As(err, &e) && e.Status != 404 ||
					!strings.Contains(e.Message, "User does not exist") ||
					!strings.Contains(e.Message, "User not found") {

					return err
				}
			}
		}
	}

	// delete invitation if exists
	for _, invitation := range pul.Invitations {
		if invitation.InvitedUserEmail == email {
			if err = client.ProjectInviteDelete(ctx, projectName, email); common.IsCritical(err) {
				return err
			}
		}
	}

	return nil
}

// isProjectUserAlreadyInvited return true if user already been invited to the project
func isProjectUserAlreadyInvited(err error) bool {
	var e avngen.Error
	if errors.As(err, &e) {
		if strings.Contains(e.Message, "already been invited to this project") && e.Status == 409 {
			return true
		}
	}

	return false
}
