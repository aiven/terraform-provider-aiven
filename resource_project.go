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
		Importer: &schema.ResourceImporter{
			State: resourceProjectState,
		},

		Schema: map[string]*schema.Schema{
			"ca_cert": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Project root CA. This is used by some services like Kafka to sign service certificate",
				Optional:    true,
			},
			"card_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Credit card ID",
			},
			"project": {
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
			CardID:  d.Get("card_id").(string),
			Project: d.Get("project").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name)
	resourceProjectGetCACert(project.Name, client, d)
	return nil
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return err
	}

	d.Set("project", project.Name)
	d.Set("card_id", project.Card.CardID)
	resourceProjectGetCACert(project.Name, client, d)
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Projects.Update(
		d.Get("project").(string),
		aiven.UpdateProjectRequest{
			CardID: d.Get("card_id").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name)
	return nil
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Projects.Delete(d.Id())
}

func resourceProjectState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("project", project.Name)
	d.Set("card_id", project.Card.CardID)
	resourceProjectGetCACert(project.Name, client, d)

	return []*schema.ResourceData{d}, nil
}

func resourceProjectGetCACert(project string, client *aiven.Client, d *schema.ResourceData) {
	ca, err := client.CA.Get(project)
	if err == nil {
		d.Set("ca_cert", ca)
	}
}
