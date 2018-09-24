// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Exists: resourceProjectExists,
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
	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return err
	}
	projectName := d.Get("project").(string)
	project, err := client.Projects.Create(
		aiven.CreateProjectRequest{
			CardID:  cardID,
			Project: projectName,
		},
	)
	if err != nil {
		return err
	}

	d.SetId(projectName)
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
	// Don't set card id unconditionally to prevent converting short card id format to long
	currentCardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil || currentCardID != project.Card.CardID {
		d.Set("card_id", project.Card.CardID)
	}
	resourceProjectGetCACert(project.Name, client, d)
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return err
	}
	project, err := client.Projects.Update(
		d.Get("project").(string),
		aiven.UpdateProjectRequest{
			CardID: cardID,
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

func resourceProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	_, err := client.Projects.Get(d.Get("project").(string))
	return resourceExists(err)
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

func getLongCardID(client *aiven.Client, cardID string) (string, error) {
	card, err := client.CardsHandler.Get(cardID)
	if err != nil {
		return "", err
	}
	if card != nil {
		return card.CardID, nil
	}
	return cardID, nil
}
