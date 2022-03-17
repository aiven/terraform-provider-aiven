package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Description: "The Project VPC data source provides information about the existing Aiven Project VPC.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectVPCSchema,
			"project", "cloud_name"),
	}
}

func datasourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, vpc := range vpcs {
		if vpc.CloudName == cloudName {
			d.SetId(schemautil.BuildResourceID(projectName, vpc.ProjectVPCID))
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
