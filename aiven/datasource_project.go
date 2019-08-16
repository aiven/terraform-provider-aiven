// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceProject() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceProjectRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenProjectSchema, "project"),
	}
}

func datasourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	project, err := client.Projects.Get(projectName)
	if err != nil {
		return err
	}

	d.SetId(projectName)

	currentCardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil || currentCardID != project.Card.CardID {
		d.Set("card_id", project.Card.CardID)
	}
	setProjectTerraformProperties(d, client, project)
	return nil
}
