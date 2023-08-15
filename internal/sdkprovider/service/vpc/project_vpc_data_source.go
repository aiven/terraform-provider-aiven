package vpc

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceProjectVPC() *schema.Resource {
	aivenProjectVPCDataSourceSchema := map[string]*schema.Schema{
		"project": {
			Type:          schema.TypeString,
			ValidateFunc:  validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "project name should be alphanumeric"),
			Description:   "Identifies the project this resource belongs to.",
			Optional:      true,
			ConflictsWith: []string{"vpc_id"},
		},
		"cloud_name": {
			Type:          schema.TypeString,
			Description:   "Defines where the cloud provider and region where the service is hosted in. See the Service resource for additional information.",
			Optional:      true,
			ConflictsWith: []string{"vpc_id"},
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
			Description: userconfig.Desc("State of the VPC.").PossibleValues("APPROVED", "ACTIVE", "DELETING", "DELETED").Build(),
		},
	}

	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Description: "The Project VPC data source provides information about the existing Aiven Project VPC.",
		Schema:      aivenProjectVPCDataSourceSchema,
	}
}

func datasourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var vpcID, projectName, cloudName string

	// This two branches are isolated by tf validation
	if s, hasID := d.GetOk("vpc_id"); hasID {
		chunks, err := schemautil.SplitResourceID(s.(string), 2)
		if err != nil {
			return diag.Errorf("error splitting vpc_id: %s:", err)
		}
		projectName = chunks[0]
		vpcID = chunks[1]
	} else {
		projectName = d.Get("project").(string)
		cloudName = d.Get("cloud_name").(string)
	}

	vpcList, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.Errorf("error getting a list of project %q VPCs: %s", projectName, err)
	}

	// At this point we have strictly either vpcID OR cloudName
	// Because of ConflictsWith: []string{"project", "cloud_name"},
	vpc, err := getVPC(vpcList, vpcID, cloudName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, vpc.ProjectVPCID))
	err = copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
	if err != nil {
		return diag.Errorf("error setting project VPC datasource values: %s", err)
	}

	return nil
}

// getVPC gets VPC by id or cloud name
func getVPC(vpcList []*aiven.VPC, vpcID, cloudName string) (vpc *aiven.VPC, err error) {
	//   A  xnor  B   | A | B | Out
	// ---------------|---|---|----
	// "foo" == ""    | 0 | 1 | 0
	//    "" == "foo" | 1 | 0 | 0
	//    "" == ""    | 1 | 1 | 1
	// "foo" == "foo" | 0 | 0 | 1
	if (vpcID == "") == (cloudName == "") {
		return nil, fmt.Errorf("provide exactly one: vpc_id or cloud_name")
	}

	for _, v := range vpcList {
		// Exact match
		if v.ProjectVPCID == vpcID {
			return v, nil
		}

		// cloudName can't be empty by this time
		if v.CloudName != cloudName {
			continue
		}

		// Cases:
		// 1. multiple active with same cloudName
		// 2. one is deleting and another one is creating (APPROVED)
		if vpc != nil {
			return nil, fmt.Errorf("multiple project VPC with cloud_name %q, use vpc_id instead", cloudName)
		}
		vpc = v
	}

	if vpc == nil {
		err = fmt.Errorf("not found project VPC with cloud_name %q", cloudName)
	}

	return vpc, err
}
