package vpc

import (
	"context"
	"strconv"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceProjectVPC() *schema.Resource {
	vpc_schema := schemautil.ResourceSchemaAsDatasourceSchema(
		aivenProjectVPCSchema,
		"project",
		"cloud_name",
	)
	vpc_schema["vpcs"] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "VPCs",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"network_cidr": {
					Required:    true,
					ForceNew:    true,
					Type:        schema.TypeString,
					Description: "Network address range used by the VPC like 192.168.0.0/24",
				},
				"state": {
					Computed:    true,
					Type:        schema.TypeString,
					Description: schemautil.Complex("State of the VPC.").PossibleValues("APPROVED", "ACTIVE", "DELETING", "DELETED").Build(),
				},
				"id": {
					Computed: true,
					Type:     schema.TypeString,
				},
			},
		},
	}
	return &schema.Resource{
		ReadContext: datasourceProjectVPCRead,
		Description: "The Project VPC data source provides information about the existing Aiven Project VPC.",
		Schema:      vpc_schema,
	}
}

func datasourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)

	vpcs_list := make([]map[string]string, 0)

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	filtered_vpcs := make([]*aiven.VPC, 0)
	for _, vpc := range vpcs {
		if vpc.CloudName == cloudName {
			filtered_vpcs = append(filtered_vpcs, vpc)
		}
	}

	if len(filtered_vpcs) == 0 {
		return diag.Errorf("project %s has no VPC defined for %s",
			projectName, cloudName)
	}

	if len(filtered_vpcs) == 1 {
		vpc := filtered_vpcs[0]
		d.SetId(schemautil.BuildResourceID(projectName, vpc.ProjectVPCID))
		err = copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	for _, vpc := range filtered_vpcs {
		data := map[string]string{
			"network_cidr": vpc.NetworkCIDR,
			"state":        vpc.State,
			"id":           schemautil.BuildResourceID(projectName, vpc.ProjectVPCID),
		}

		vpcs_list = append(vpcs_list, data)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cloud_name", cloudName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("vpcs", vpcs_list); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
