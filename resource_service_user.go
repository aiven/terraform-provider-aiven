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
		Delete: resourceServiceUserDelete,

		Schema: map[string]*schema.Schema{
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
				Description: "Service username",
				ForceNew:    true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"access_cert": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"access_key": {
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
			Username: d.Get("username").(string),
		},
	)
	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(user.Username + "!")

	d.Set("username", user.Username)
	d.Set("password", user.Password)
	d.Set("type", user.Type)
	d.Set("access_cert", user.AccessCert)
	d.Set("access_key", user.AccessKey)

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
			d.Set("access_cert", user.AccessCert)
			d.Set("access_key", user.AccessKey)
			return nil
		}
	}

	return errors.New("User not found")
}

func resourceServiceUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.ServiceUsers.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("username").(string),
	)
}
