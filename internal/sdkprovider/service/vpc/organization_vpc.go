package vpc

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationVPCSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The ID of the organization.",
	},
	"cloud_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The cloud provider and region where the service is hosted in the format `CLOUD_PROVIDER-REGION_NAME`. For example, `google-europe-west1` or `aws-us-east-2`.").ForceNew().Build(),
	},
	"network_cidr": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.IsCIDR,
		Description:  "Network address range used by the VPC. For example, `192.168.0.0/24`.",
	},
	"organization_vpc_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the Aiven Organization VPC.",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: userconfig.Desc("State of the VPC.").PossibleValuesString(organizationvpc.VpcStateTypeChoices()...).Build(),
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation of the VPC.",
	},
	"update_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of the last update of the VPC.",
	},
}

func ResourceOrganizationVPC() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a VPC for an Aiven organization.",
		CreateContext: common.WithGenClient(resourceOrganizationVPCCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationVPCRead),
		DeleteContext: common.WithGenClient(resourceOrganizationVPCDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationVPCSchema,
	}
}

func resourceOrganizationVPCCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID = d.Get("organization_id").(string)
		cloud = d.Get("cloud_name").(string)
		cidr  = d.Get("network_cidr").(string)
	)

	resp, err := client.OrganizationVpcCreate(ctx, orgID, &organizationvpc.OrganizationVpcCreateIn{
		Clouds: []organizationvpc.CloudIn{
			{
				CloudName:   cloud,
				NetworkCidr: cidr,
			},
		},
		PeeringConnections: make([]organizationvpc.PeeringConnectionIn, 0), // nil here would cause an error from the API
	})
	if err != nil {
		return err
	}

	// Wait for VPC to become active
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(organizationvpc.VpcStateTypeApproved)},
		Target:  []string{string(organizationvpc.VpcStateTypeActive)},
		Refresh: func() (interface{}, string, error) {
			orgVPC, err := client.OrganizationVpcGet(ctx, orgID, resp.OrganizationVpcId)
			if err != nil {
				return nil, "", err
			}

			return orgVPC, string(orgVPC.State), nil
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for VPC (%q) to become active: %w", resp.OrganizationVpcId, err)
	}

	d.SetId(schemautil.BuildResourceID(orgID, resp.OrganizationVpcId))

	return resourceOrganizationVPCRead(ctx, d, client)
}

func resourceOrganizationVPCRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, vpcID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
	if err != nil {
		return err
	}

	// currently we support only 1 cloud per VPC
	if len(resp.Clouds) != 1 {
		return fmt.Errorf("expected exactly 1 cloud, got %d", len(resp.Clouds))
	}

	if err = d.Set("organization_id", orgID); err != nil {
		return err
	}
	if err = d.Set("organization_vpc_id", resp.OrganizationVpcId); err != nil {
		return err
	}
	if err = d.Set("cloud_name", resp.Clouds[0].CloudName); err != nil {
		return err
	}
	if err = d.Set("network_cidr", resp.Clouds[0].NetworkCidr); err != nil {
		return err
	}
	if err = d.Set("state", resp.State); err != nil {
		return err
	}
	if err = d.Set("create_time", resp.CreateTime.String()); err != nil {
		return err
	}
	if err = d.Set("update_time", resp.UpdateTime.String()); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationVPCDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, vpcID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	_, err = client.OrganizationVpcDelete(ctx, orgID, vpcID)
	if common.IsCritical(err) {
		return err
	}

	// Wait for VPC to be deleted
	stateConf := &retry.StateChangeConf{
		Target: []string{string(organizationvpc.VpcStateTypeDeleted)},
		Refresh: func() (interface{}, string, error) {
			orgVPC, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
			if err != nil {
				if avngen.IsNotFound(err) {
					// if resource is not found, it's considered deleted
					// return empty struct instead of nil to avoid the "not found" counter behavior in StateChangeConf when Target states are specified
					return struct{}{}, string(organizationvpc.VpcStateTypeDeleted), nil
				}

				return nil, "", err
			}

			return orgVPC, string(orgVPC.State), nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if common.IsCritical(err) {
		return err
	}

	return nil
}
