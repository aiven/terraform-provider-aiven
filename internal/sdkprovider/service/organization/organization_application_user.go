package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationApplicationUserSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Description: "The ID of the organization the application user belongs to.",
		Required:    true,
		ForceNew:    true,
	},
	"name": {
		Type:        schema.TypeString,
		Description: "Name of the application user.",
		Required:    true,
	},
	"is_super_admin": {
		Type: schema.TypeBool,
		Description: "Makes the application user a super admin. The super admin role has full access to an organization, " +
			"its billing and settings, and all its organizational units, projects, and services.",
		Optional: true,
	},
	"user_id": {
		Type:        schema.TypeString,
		Description: "The ID of the application user.",
		Computed:    true,
	},
	"email": {
		Type:        schema.TypeString,
		Description: `An email address automatically generated by Aiven to help identify the application user. No notifications are sent to this email.`,
		Computed:    true,
	},
}

func ResourceOrganizationApplicationUser() *schema.Resource {
	return &schema.Resource{
		Description: userconfig.Desc(`
Creates and manages an organization application user. [Application users](https://aiven.io/docs/platform/concepts/application-users) can be used for
programmatic access to the platform.

You give application users access to projects by adding them as members of a group using ` + "`aiven_organization_user_group_member`" + `
and assigning the group to a project with ` + "`aiven_organization_group_project`" + `. You can give an application user access to all
resources in your organization by setting ` + "`is_super_admin = true`" + ` .`,
		).Build(),
		CreateContext: common.WithGenClient(resourceOrganizationApplicationUserCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationApplicationUserRead),
		UpdateContext: common.WithGenClient(resourceOrganizationApplicationUserUpdate),
		DeleteContext: common.WithGenClient(resourceOrganizationApplicationUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOrganizationApplicationUserSchema,
	}
}

func resourceOrganizationApplicationUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req applicationuser.ApplicationUserCreateIn
	err := schemautil.ResourceDataGet(d, &req)
	if err != nil {
		return err
	}

	orgID := d.Get("organization_id").(string)
	user, err := client.ApplicationUserCreate(ctx, orgID, &req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(orgID, user.UserId))
	return resourceOrganizationApplicationUserRead(ctx, d, client)
}

func resourceOrganizationApplicationUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, usrID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	// First gets app user info, then regular user info
	users, err := client.ApplicationUsersList(ctx, orgID)
	if err != nil {
		return err
	}

	var user *applicationuser.ApplicationUserOut
	for i := range users {
		if usrID == users[i].UserId {
			user = &users[i]
			break
		}
	}

	if user == nil {
		return fmt.Errorf(`application user not found`)
	}

	// Sets name and user_id
	err = schemautil.ResourceDataSet(aivenOrganizationApplicationUserSchema, d, user)
	if err != nil {
		return err
	}

	// This field has "email" in the schema and "user_email" in the request
	if err = d.Set("email", user.UserEmail); err != nil {
		return err
	}

	// This is for import command
	if err = d.Set("organization_id", orgID); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationApplicationUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	var req applicationuser.ApplicationUserUpdateIn
	err = schemautil.ResourceDataGet(d, &req)
	if err != nil {
		return err
	}

	_, err = client.ApplicationUserUpdate(ctx, orgID, userID, &req)
	if err != nil {
		return err
	}

	return resourceOrganizationApplicationUserRead(ctx, d, client)
}

func resourceOrganizationApplicationUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	err = client.ApplicationUserDelete(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete application user %s/%s: %w", orgID, userID, err)
	}
	return nil
}
