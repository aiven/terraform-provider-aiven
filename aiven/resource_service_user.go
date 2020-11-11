// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenServiceUserSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the user to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the user to",
		ForceNew:    true,
	},
	"username": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the user account",
		ForceNew:    true,
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account",
	},
	"password": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Password of the user",
	},
	"access_cert": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate for the user if applicable for the service in question",
	},
	"access_key": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate key for the user if applicable for the service in question",
	},
}

func resourceServiceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceUserCreate,
		Read:   resourceServiceUserRead,
		Delete: resourceServiceUserDelete,
		Exists: resourceServiceUserExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceUserState,
		},

		Schema: aivenServiceUserSchema,
	}
}

func resourceServiceUserCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	user, err := client.ServiceUsers.Create(
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
		},
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName, username))
	return copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
}

func copyServiceUserPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	user *aiven.ServiceUser,
	projectName string,
	serviceName string,
) error {
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("username", user.Username); err != nil {
		return err
	}
	if err := d.Set("password", user.Password); err != nil {
		return err
	}
	if err := d.Set("type", user.Type); err != nil {
		return err
	}
	if err := d.Set("access_cert", user.AccessCert); err != nil {
		return err
	}
	if err := d.Set("access_key", user.AccessKey); err != nil {
		return err
	}

	return nil
}

func resourceServiceUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())
	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return err
	}

	return copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
}

func resourceServiceUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())
	return client.ServiceUsers.Delete(projectName, serviceName, username)
}

func resourceServiceUserExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())
	_, err := client.ServiceUsers.Get(projectName, serviceName, username)
	return resourceExists(err)
}

func resourceServiceUserState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<username>", d.Id())
	}

	projectName, serviceName, username := splitResourceID3(d.Id())
	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return nil, err
	}

	err = copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
