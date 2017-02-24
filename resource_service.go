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
			"service_name": &schema.Schema{Type: schema.TypeString,
				Required:    true,
				Description: "Service name",
			},
			"service_type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service type code",
			},
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Service type code",
			},
		},
	}
}

func resourceServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Services.Create(
		d.Get("project").(string),
		aiven.CreateServiceRequest{
			d.Get("cloud").(string),
			d.Get("group_name").(string),
			d.Get("plan").(string),
			d.Get("service_name").(string),
			d.Get("service_type").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name + "!")
	return nil
}

func resourceServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Get(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
	if err != nil {
		return err
	}

	d.Set("name", service.Name)
	d.Set("host", service.Uri)

	return nil
}

func resourceServiceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Update(
		d.Get("project").(string),
		d.Get("service_name").(string),
		aiven.UpdateServiceRequest{
			d.Get("cloud").(string),
			d.Get("group_name").(string),
			d.Get("plan").(string),
			true,
		},
	)
	if err != nil {
		return err
	}

	d.Set("name", service.Name)
	return nil
}

func resourceServiceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Services.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
}
