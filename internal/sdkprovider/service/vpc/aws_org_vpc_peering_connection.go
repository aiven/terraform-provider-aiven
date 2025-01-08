package vpc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAWSOrgVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Identifier of the organization.",
	},
	"organization_vpc_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Identifier of the organization VPC.",
	},
	"aws_account_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("AWS account ID.").ForceNew().Build(),
	},
	"aws_vpc_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("AWS VPC ID.").ForceNew().Build(),
	},
	"aws_vpc_region": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The AWS region of the peered VPC. For example, `eu-central-1`.").Build(),
	},
	"peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The ID of the peering connection.").Build(),
	},
	"aws_vpc_peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "The ID of the AWS VPC peering connection.",
	},
	"state": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("State of the peering connection.").PossibleValuesString(organizationvpc.VpcPeeringConnectionStateTypeChoices()...).Build(),
	},
}

func ResourceAWSOrgVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an AWS VPC peering connection with an Aiven Organization VPC.",
		CreateContext: common.WithGenClientDiag(resourceAWSOrgVPCPeeringConnectionCreate),
		ReadContext:   common.WithGenClientDiag(resourceAWSOrgVPCPeeringConnectionRead),
		DeleteContext: common.WithGenClientDiag(resourceAWSOrgVPCPeeringConnectionDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAWSOrgVPCPeeringConnectionSchema,
	}
}

func resourceAWSOrgVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	var (
		orgID        = d.Get("organization_id").(string)
		vpcID        = d.Get("organization_vpc_id").(string)
		awsAccountID = d.Get("aws_account_id").(string)
		awsVPCId     = d.Get("aws_vpc_id").(string)
		awsRegion    = d.Get("aws_vpc_region").(string)

		req = organizationvpc.OrganizationVpcPeeringConnectionCreateIn{
			PeerRegion:       util.ToPtr(awsRegion),
			PeerVpc:          awsVPCId,
			PeerCloudAccount: awsAccountID,
		}
	)

	pCon, err := createPeeringConnection(ctx, orgID, vpcID, client, d, req)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	diags := getDiagnosticsFromState(newOrganizationVPCPeeringState(pCon))

	d.SetId(schemautil.BuildResourceID(orgID, vpcID, awsAccountID, awsVPCId, awsRegion))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		deleteDiags := resourceAzureOrgVPCPeeringConnectionDelete(ctx, d, client)
		d.SetId("") // Clear the ID after delete

		return append(diags, deleteDiags...)
	}

	return append(diags, resourceAWSOrgVPCPeeringConnectionRead(ctx, d, client)...)
}

func resourceAWSOrgVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	orgID, vpcID, awsAccountID, awsVpcID, awsRegion, err := schemautil.SplitResourceID5(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	vpc, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
	if err != nil {
		if avngen.IsNotFound(err) {
			return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
		}

		return diag.Errorf("failed to get VPC with ID %q: %s", vpcID, err)
	}

	pc := lookupAWSPeeringConnection(vpc, awsAccountID, awsVpcID, awsRegion)
	if pc == nil {
		d.SetId("") // Clear the ID as the resource is not found

		return diag.FromErr(fmt.Errorf("VPC peering connection not found"))
	}

	if err = d.Set("organization_id", vpc.OrganizationId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("organization_vpc_id", vpc.OrganizationVpcId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("peering_connection_id", *pc.PeeringConnectionId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("aws_account_id", pc.PeerCloudAccount); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("aws_vpc_id", pc.PeerVpc); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("aws_vpc_region", *pc.PeerRegion); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("aws_vpc_peering_connection_id", pc.StateInfo.AwsVpcPeeringConnectionId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("state", string(pc.State)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAWSOrgVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	orgID, vpcID, awsAccountID, awsVpcID, awsRegion, err := schemautil.SplitResourceID5(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	vpc, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
	if err != nil {
		if avngen.IsNotFound(err) {
			return nil // consider already deleted
		}

		return diag.Errorf("failed to get VPC with ID %q: %s", vpcID, err)
	}

	if err = deletePeeringConnection(
		ctx,
		orgID,
		vpcID,
		client,
		d,
		lookupAWSPeeringConnection(vpc, awsAccountID, awsVpcID, awsRegion),
	); err != nil {
		return diag.Errorf("Error deleting Azure Aiven VPC Peering Connection: %s", err)
	}

	return nil
}

func lookupAWSPeeringConnection(
	vpc *organizationvpc.OrganizationVpcGetOut,
	awsAccountID, awsVpcID, awsRegion string,
) *organizationvpc.OrganizationVpcGetPeeringConnectionOut {
	for _, pCon := range vpc.PeeringConnections {
		if pCon.PeerCloudAccount == awsAccountID &&
			pCon.PeerVpc == awsVpcID &&
			pCon.PeerRegion != nil &&
			*pCon.PeerRegion == awsRegion &&
			pCon.PeeringConnectionId != nil {

			return &pCon
		}
	}

	return nil
}
