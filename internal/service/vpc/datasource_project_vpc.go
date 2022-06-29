package vpc

import (
	"context"
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceProjectVPC() *schema.Resource {
	aivenProjectVPCDataSourceSchema := schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectVPCSchema,
		"project", "cloud_name")
	aivenProjectVPCDataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "ID of the VPC. This can be used to filter out the specific VPC if there are more than one datasource returned.",
		Optional:    true,
	}

	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Description: "The Project VPC data source provides information about the existing Aiven Project VPC.",
		Schema:      aivenProjectVPCDataSourceSchema,
	}
}

func datasourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)
	vpcId, hasId := d.GetOk("id")

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	filteredVPCs := make([]*aiven.VPC, 0)

	for _, vpc := range vpcs {
		condition := vpc.CloudName == cloudName

		// We don't need to care about the cloud name if the ID is provided
		if hasId {
			splitId := schemautil.SplitResourceID(vpcId.(string), 2)
			if len(splitId) < 2 {
				return diag.Errorf("Invalid VPC id %s", vpcId.(string))
			}

			id := splitId[1]
			condition = vpc.ProjectVPCID == id
		}

		if condition {
			filteredVPCs = append(filteredVPCs, vpc)
		}
	}

	if len(filteredVPCs) == 0 {
		return diag.Errorf("project %s has no VPC defined for %s",
			projectName, cloudName)
	}

	if len(filteredVPCs) > 1 {
		// List out the available options in the error message
		var smg string
		for _, vpc := range filteredVPCs {
			id := schemautil.BuildResourceID(projectName, vpc.ProjectVPCID)
			smg = smg + fmt.Sprintf("- ID=(%v), State=(%v), NetworkCIDR=(%v)\n", id, vpc.State, vpc.NetworkCIDR)
		}
		return diag.Errorf("project %s has multiple VPC defined for %s. Please add `id` to get the desired one. The available vpc ids are:\n%s",
			projectName, cloudName, smg)
	}

	vpc := filteredVPCs[0]
	d.SetId(schemautil.BuildResourceID(projectName, vpc.ProjectVPCID))
	err = copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
