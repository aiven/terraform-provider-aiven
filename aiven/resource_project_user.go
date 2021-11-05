// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenProjectUserSchema = map[string]*schema.Schema{
	"project": commonSchemaProjectReference,
	"email": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: complex("Email address of the user.").forceNew().build(),
	},
	"member_type": {
		Required:    true,
		Type:        schema.TypeString,
		Description: complex("Project membership type.").possibleValues("admin", "developer", "operator").build(),
	},
	"accepted": {
		Computed:    true,
		Type:        schema.TypeBool,
		Description: "Whether the user has accepted the request to join the project; adding user to a project sends an invitation to the target user and the actual membership is only created once the user accepts the invitation.",
	},
}

func resourceProjectUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The Project User resource allows the creation and management of Aiven Project Users.",
		CreateContext: resourceProjectUserCreate,
		ReadContext:   resourceProjectUserRead,
		UpdateContext: resourceProjectUserUpdate,
		DeleteContext: resourceProjectUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectUserState,
		},

		Schema: aivenProjectUserSchema,
	}
}

func resourceProjectUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	email := d.Get("email").(string)
	err := client.ProjectUsers.Invite(
		projectName,
		aiven.CreateProjectInvitationRequest{
			UserEmail:  email,
			MemberType: d.Get("member_type").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(projectName, email))
	if err := d.Set("accepted", false); err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectUserRead(ctx, d, m)
}

func resourceProjectUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		if aiven.IsNotFound(err) && !d.Get("accepted").(bool) {
			return resourceProjectUserCreate(ctx, d, m)
		}
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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

	projectName, email := splitResourceID2(d.Id())
	memberType := d.Get("member_type").(string)
	err := client.ProjectUsers.UpdateUserOrInvitation(
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

func resourceProjectUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		return diag.FromErr(err)
	}

	// delete user if exists
	if user != nil {
		err := client.ProjectUsers.DeleteUser(projectName, email)
		if err != nil {
			if err.(aiven.Error).Status != 404 ||
				!strings.Contains(err.(aiven.Error).Message, "User does not exist") ||
				!strings.Contains(err.(aiven.Error).Message, "User not found") {

				return diag.FromErr(err)
			}
		}
	}

	// delete invitation if exists
	if invitation != nil {
		err := client.ProjectUsers.DeleteInvitation(projectName, email)
		if err != nil && !aiven.IsNotFound(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceProjectUserState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<email>", d.Id())
	}

	di := resourceProjectUserRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get project user %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
