// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var aivenProjectUserSchema = map[string]*schema.Schema{
	"project": {
		Description: "The project the user belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"email": {
		Description: "Email address of the user",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"member_type": {
		Description: "Project membership type. One of: admin, developer, operator",
		Required:    true,
		Type:        schema.TypeString,
	},
	"accepted": {
		Computed:    true,
		Description: "Whether the user has accepted project membership or not",
		Type:        schema.TypeBool,
	},
}

func resourceProjectUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectUserCreate,
		Read:   resourceProjectUserRead,
		Update: resourceProjectUserUpdate,
		Delete: resourceProjectUserDelete,
		Exists: resourceProjectUserExists,
		Importer: &schema.ResourceImporter{
			State: resourceProjectUserState,
		},

		Schema: aivenProjectUserSchema,
	}
}

func resourceProjectUserCreate(d *schema.ResourceData, m interface{}) error {
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
		return err
	}

	d.SetId(buildResourceID(projectName, email))
	if err := d.Set("accepted", false); err != nil {
		return err
	}
	return nil
}

func resourceProjectUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		return err
	}

	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("email", email); err != nil {
		return err
	}
	if user != nil {
		if err := d.Set("member_type", user.MemberType); err != nil {
			return err
		}
		if err := d.Set("accepted", true); err != nil {
			return err
		}
	} else {
		if err := d.Set("member_type", invitation.MemberType); err != nil {
			return err
		}
		if err := d.Set("accepted", false); err != nil {
			return err
		}
	}
	return nil
}

func resourceProjectUserUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	memberType := d.Get("member_type").(string)
	return client.ProjectUsers.UpdateUserOrInvitation(
		projectName,
		email,
		aiven.UpdateProjectUserOrInvitationRequest{
			MemberType: memberType,
		},
	)
}

func resourceProjectUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		return err
	}

	// delete user if exists
	if user != nil {
		err := client.ProjectUsers.DeleteUser(projectName, email)
		if err != nil {
			if err.(aiven.Error).Status != 404 ||
				!strings.Contains(err.(aiven.Error).Message, "User does not exist") ||
				!strings.Contains(err.(aiven.Error).Message, "User not found") {

				return err
			}
		}
	}

	// delete invitation if exists
	if invitation != nil {
		err := client.ProjectUsers.DeleteInvitation(projectName, email)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}
	}

	return nil
}

func resourceProjectUserExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	_, _, err := client.ProjectUsers.Get(projectName, email)
	return resourceExists(err)
}

func resourceProjectUserState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<email>", d.Id())
	}

	err := resourceProjectUserRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
