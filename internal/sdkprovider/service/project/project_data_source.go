package project

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceProjectRead),
		Description: "Gets information about an Aiven project.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectSchema, "project"),
	}
}

func datasourceProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)

	resp, err := client.ProjectList(ctx)
	if err != nil {
		return err
	}

	for _, project := range resp.Projects {
		if project.ProjectName == projectName {
			d.SetId(projectName)

			return resourceProjectRead(ctx, d, client)
		}
	}

	return fmt.Errorf("project %s not found", projectName)
}
