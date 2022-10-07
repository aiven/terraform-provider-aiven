package vpc

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceProjectVPC() *schema.Resource {
	aivenProjectVPCDataSourceSchema := map[string]*schema.Schema{
		"project": {
			Type:          schema.TypeString,
			ValidateFunc:  validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "project name should be alphanumeric"),
			Description:   "Identifies the project this resource belongs to.",
			Optional:      true,
			ConflictsWith: []string{"vpc_id"},
			Deprecated:    "Use vpc_id instead of project/cloud_name. Only vpc_id can correctly retrieve a project VPC.",
		},
		"cloud_name": {
			Type:          schema.TypeString,
			Description:   "Defines where the cloud provider and region where the service is hosted in. See the Service resource for additional information.",
			Optional:      true,
			ConflictsWith: []string{"vpc_id"},
			Deprecated:    "Use vpc_id instead of project/cloud_name. Only vpc_id can correctly retrieve a project VPC.",
		},
		"vpc_id": {
			Type:          schema.TypeString,
			Description:   "ID of the VPC. This can be used to filter out the specific VPC if there are more than one datasource returned.",
			Optional:      true,
			ConflictsWith: []string{"project", "cloud_name"},
			ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
				_, err := schemautil.SplitResourceID(i.(string), 2)
				if err != nil {
					return diag.Errorf("invalid vpc_id, should have the following format {project_name}/{project_vpc_id}: %s", err)
				}
				return nil
			},
		},
		"network_cidr": {
			Computed:    true,
			Type:        schema.TypeString,
			Description: "Network address range used by the VPC like 192.168.0.0/24",
		},
		"state": {
			Computed:    true,
			Type:        schema.TypeString,
			Description: schemautil.Complex("State of the VPC.").PossibleValues("APPROVED", "ACTIVE", "DELETING", "DELETED").Build(),
		},
	}

	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Description: "The Project VPC data source provides information about the existing Aiven Project VPC.",
		Schema:      aivenProjectVPCDataSourceSchema,
	}
}

func datasourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var splitId []string
	var vpcId string
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)
	if id, hasId := d.GetOk("vpc_id"); hasId {
		var err error
		splitId, err = schemautil.SplitResourceID(id.(string), 2)
		if err != nil {
			return diag.Errorf("error splitting vpc_id: %s:", err)
		}

		projectName = splitId[0]
		vpcId = splitId[1]
	}

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.Errorf("error getting a list of project VPCs: %s", err)
	}

	filteredVPCs := make([]*aiven.VPC, 0)
	for _, vpc := range vpcs {
		if vpc.CloudName == cloudName || (vpcId != "" && vpc.ProjectVPCID == vpcId) {
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
		return diag.Errorf("error setting project VPC datasource values: %s", err)
	}

	return nil
}
