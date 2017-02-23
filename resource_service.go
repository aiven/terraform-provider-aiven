package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,

		// TODO: add user config
		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target cloud",
			}, "cloud": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target cloud",
			},
			"group_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service group name",
			},
			"plan": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subscription plan",
			},
			"service_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service name",
			},
			"service_type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service type code",
			},
		},
	}
}

func resourceServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Services.Create(
		d.Get("project").(string),
		d.Get("cloud").(string),
		d.Get("group_name").(string),
		d.Get("plan").(string),
		d.Get("service_name").(string),
		d.Get("service_type").(string),
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name + "!")
	return nil
}

func resourceServiceRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServiceUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServiceDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
