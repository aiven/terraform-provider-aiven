package project

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceProjectRead,
		Description: "The Project data source provides information about the existing Aiven Project.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectSchema, "project"),
	}
}

func datasourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)

	projects, err := client.Projects.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, project := range projects {
		if project.Name == projectName {
			d.SetId(projectName)
			return resourceProjectRead(ctx, d, m)
		}
	}

	return diag.Errorf("project %s not found", projectName)
}
