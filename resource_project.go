package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,

		Schema: map[string]*schema.Schema{
			"card_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Credit card ID",
			},
			"cloud": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target cloud",
			},
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project name",
			},
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	project, err := client.Projects.Create(
		aiven.CreateProjectRequest{
			d.Get("card_id").(string),
			d.Get("cloud").(string),
			d.Get("project").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name + "!")
	return nil
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Get("project").(string))
	if err != nil {
		return err
	}

	d.Set("project", project.Name)
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Projects.Update(
		d.Get("project").(string),
		aiven.UpdateProjectRequest{
			d.Get("card_id").(string),
			d.Get("cloud").(string),
			d.Get("project").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name + "!")
	return nil
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Projects.Delete(d.Get("project").(string))
}
