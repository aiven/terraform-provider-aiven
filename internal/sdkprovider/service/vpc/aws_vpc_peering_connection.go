package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAWSVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		Description:  userconfig.Desc("The ID of the Aiven VPC.").ForceNew().Build(),
		ValidateFunc: validateVPCID,
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
		ForceNew: true,
		Required: true,
		Type:     schema.TypeString,
		DiffSuppressFunc: func(_, _, newValue string, _ *schema.ResourceData) bool {
			return newValue == ""
		},
		Description: userconfig.Desc("The AWS region of the peered VPC, if different from the Aiven VPC region.").ForceNew().Build(),
	},
	"state": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "The state of the peering connection.",
	},
	"state_info": {
		Computed:    true,
		Type:        schema.TypeMap,
		Description: "State-specific help or error information.",
	},
	"aws_vpc_peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "The ID of the AWS VPC peering connection.",
	},
}

func ResourceAWSVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an AWS VPC peering connection with an Aiven VPC.",
		CreateContext: resourceAWSVPCPeeringConnectionCreate,
		ReadContext:   resourceAWSVPCPeeringConnectionRead,
		DeleteContext: resourceAWSVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAWSVPCPeeringConnectionSchema,
	}
}

func resourceAWSVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
	)

	client := m.(*aiven.Client)
	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	awsAccountID := d.Get("aws_account_id").(string)
	awsVPCId := d.Get("aws_vpc_id").(string)
	awsVPCRegion := d.Get("aws_vpc_region").(string)
	if awsVPCRegion != "" {
		region = &awsVPCRegion
	}

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		ctx,
		projectName,
		vpcID,
		awsAccountID,
		awsVPCId,
		&awsVPCRegion,
	)
	if common.IsCritical(err) {
		return diag.Errorf("error checking aws peering connection: %s", err)
	}

	if pc != nil {
		return diag.Errorf("aws vpc peering connection already exists and cannot be created")
	}

	if _, err = client.VPCPeeringConnections.Create(
		ctx,
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount: awsAccountID,
			PeerVPC:          awsVPCId,
			PeerRegion:       region,
		},
	); err != nil {
		return diag.Errorf("Error waiting for AWS VPC peering connection creation: %s", err)
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{"APPROVED"},
		Target: []string{
			"ACTIVE",
			"REJECTED_BY_PEER",
			"PENDING_PEER",
			"INVALID_SPECIFICATION",
			"DELETING",
			"DELETED",
			"DELETED_BY_PEER",
		},
		Refresh: func() (interface{}, string, error) {
			pc, err := client.VPCPeeringConnections.GetVPCPeering(
				ctx,
				projectName,
				vpcID,
				awsAccountID,
				awsVPCId,
				region,
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}

	res, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	pc = res.(*aiven.VPCPeeringConnection)
	diags := getDiagnosticsFromState(newAivenVPCPeeringState(pc))

	d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		return append(diags, resourceAWSVPCPeeringConnectionDelete(ctx, d, m)...)
	}

	return append(diags, resourceAWSVPCPeeringConnectionRead(ctx, d, m)...)
}

func resourceAWSVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing AWS peering VPC ID: %s", err)
	}

	pc, err := client.VPCPeeringConnections.GetVPCPeering(
		ctx,
		p.projectName,
		p.vpcID,
		p.peerCloudAccount,
		p.peerVPC,
		p.peerRegion,
	)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	return copyAWSVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, p.projectName, p.vpcID)
}

func resourceAWSVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing AWS peering VPC ID: %s", err)
	}

	err = client.VPCPeeringConnections.DeleteVPCPeering(
		ctx,
		p.projectName,
		p.vpcID,
		p.peerCloudAccount,
		p.peerVPC,
		p.peerRegion,
	)
	if common.IsCritical(err) {
		return diag.Errorf("Error deleting VPC peering connection: %s", err)
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{
			"ACTIVE",
			"APPROVED",
			"APPROVED_PEER_REQUESTED",
			"DELETING",
			"INVALID_SPECIFICATION",
			"PENDING_PEER",
			"REJECTED_BY_PEER",
			"DELETED_BY_PEER",
		},
		Target: []string{
			"DELETED",
		},
		Refresh: func() (interface{}, string, error) {
			pc, err := client.VPCPeeringConnections.GetVPCPeering(
				ctx,
				p.projectName,
				p.vpcID,
				p.peerCloudAccount,
				p.peerVPC,
				p.peerRegion,
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error waiting for AWS Aiven VPC Peering Connection to be DELETED: %s", err)
	}
	return nil
}

func copyAWSVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) diag.Diagnostics {
	if err := d.Set("vpc_id", schemautil.BuildResourceID(project, vpcID)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_account_id", peeringConnection.PeerCloudAccount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_vpc_id", peeringConnection.PeerVPC); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", peeringConnection.State); err != nil {
		return diag.FromErr(err)
	}

	if peeringConnection.StateInfo != nil {
		peeringID, ok := (*peeringConnection.StateInfo)["aws_vpc_peering_connection_id"]
		if ok {
			if err := d.Set("aws_vpc_peering_connection_id", peeringID); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if peeringConnection.PeerRegion != nil {
		if err := d.Set("aws_vpc_region", peeringConnection.PeerRegion); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("state_info", ConvertStateInfoToMap(peeringConnection.StateInfo)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
