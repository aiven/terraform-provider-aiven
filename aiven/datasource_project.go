// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	projects, err := client.Projects.List()
	if err != nil {
		return err
	}

	for _, project := range projects {
		if project.Name == projectName {
			d.SetId(projectName)
			return resourceProjectRead(d, m)
		}
	}

	return fmt.Errorf("project %s not found", projectName)
}
