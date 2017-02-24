package main

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceServiceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceUserCreate,
		Read:   resourceServiceUserRead,
		Update: resourceServiceUserUpdate,
		Delete: resourceServiceUserDelete,

		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project to link the service to",
			},
			"service_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to link the service to",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service username",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
		},
	}
}

func resourceServiceUserCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	user, err := client.ServiceUsers.Create(
		d.Get("project").(string),
		d.Get("service_name").(string),
		aiven.CreateServiceUserRequest{
			d.Get("username").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(user.Username + "!")

	d.Set("username", user.Username)
	d.Set("password", user.Password)
	d.Set("type", user.Type)

	return nil
}

func resourceServiceUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Get(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
	if err != nil {
		return err
	}

	username := d.Get("username").(string)
	for _, user := range service.Users {
		if user.Username == username {
			d.Set("username", user.Username)
			d.Set("password", user.Password)
			d.Set("type", user.Type)
			return nil
		}
	}

	return errors.New("User not found")
}

func resourceServiceUserUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServiceUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.ServiceUsers.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("username").(string),
	)
}
