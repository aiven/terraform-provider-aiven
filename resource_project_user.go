package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

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

		Schema: map[string]*schema.Schema{
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
				Description: "Project membership type",
				Required:    true,
				Type:        schema.TypeString,
			},
			"accepted": {
				Computed:    true,
				Description: "Whether the user has accepted project membership or not",
				Type:        schema.TypeBool,
			},
		},
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
	d.Set("accepted", false)
	return nil
}

func resourceProjectUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	user, invitation, err := client.ProjectUsers.Get(projectName, email)
	if err != nil {
		return err
	}

	d.Set("project", projectName)
	d.Set("email", email)
	if user != nil {
		d.Set("member_type", user.MemberType)
		d.Set("accepted", true)
	} else {
		d.Set("member_type", invitation.MemberType)
		d.Set("accepted", false)
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
	return client.ProjectUsers.DeleteUserOrInvitation(projectName, email)
}

func resourceProjectUserExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, email := splitResourceID2(d.Id())
	_, _, err := client.ProjectUsers.Get(projectName, email)
	return resourceExists(err)
}

func resourceProjectUserState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<email>", d.Id())
	}

	err := resourceProjectUserRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
