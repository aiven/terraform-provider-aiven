// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenProjectVPCSchema,
			"project", "cloud_name"),
	}
}

func datasourceProjectVPCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, vpc := range vpcs {
		if vpc.CloudName == cloudName {
			d.SetId(buildResourceID(projectName, vpc.ProjectVPCID))
			err = copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
			if err != nil {
				return diag.FromErr(err)
			}

			return nil
		}
	}

	return diag.Errorf("project %s has no VPC defined for %s",
		projectName, cloudName)
}