package project

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceProjectRead,
		Description: "The Project data source provides information about the existing Aiven Project.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectSchema, "project"),
	}
}

func datasourceProjectRead(c context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)

	projects, err := client.Projects.List()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, project := range projects {
		if project.Name == projectName {
			d.SetId(projectName)

			return resourceProjectRead(c, d, m)
		}
	}

	return diag.Errorf("project %s not found", projectName)
}
