package vpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenAWSVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		Description:  schemautil.Complex("The VPC the peering connection belongs to.").ForceNew().Build(),
		ValidateFunc: validateVPCID,
	},
	"aws_account_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("AWS account ID.").ForceNew().Build(),
	},
	"aws_vpc_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("AWS VPC ID.").ForceNew().Build(),
	},
	"aws_vpc_region": {
		ForceNew: true,
		Required: true,
		Type:     schema.TypeString,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == ""
		},
		Description: schemautil.Complex("AWS region of the peered VPC (if not in the same region as Aiven VPC).").ForceNew().Build(),
	},
	"state": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "State of the peering connection",
	},
	"state_info": {
		Computed:    true,
		Type:        schema.TypeMap,
		Description: "State-specific help or error information",
	},
	"aws_vpc_peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "AWS VPC peering connection ID",
	},
}

func ResourceAWSVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "The AWS VPC Peering Connection resource allows the creation and management of Aiven AWS VPC Peering Connections.",
		CreateContext: resourceAWSVPCPeeringConnectionCreate,
		ReadContext:   resourceAWSVPCPeeringConnectionRead,
		DeleteContext: resourceAWSVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAWSVPCPeeringConnectionImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenAWSVPCPeeringConnectionSchema,
	}
}

func resourceAWSVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
	)

	client := m.(*meta.Meta).Client
	projectName, vpcID := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	awsAccountId := d.Get("aws_account_id").(string)
	awsVPCId := d.Get("aws_vpc_id").(string)
	awsVPCRegion := d.Get("aws_vpc_region").(string)
	if awsVPCRegion != "" {
		region = &awsVPCRegion
	}

	if _, err = client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount: awsAccountId,
			PeerVPC:          awsVPCId,
			PeerRegion:       region,
		},
	); err != nil {
		return diag.Errorf("Error waiting for AWS VPC peering connection creation: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
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
				projectName,
				vpcID,
				awsAccountId,
				awsVPCId,
				region,
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 2 * time.Second,
	}

	res, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	pc = res.(*aiven.VPCPeeringConnection)
	diags := getDiagnosticsFromState(pc)

	d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		return append(diags, resourceAWSVPCPeeringConnectionDelete(ctx, d, m)...)
	}

	return append(diags, resourceAWSVPCPeeringConnectionRead(ctx, d, m)...)
}

func resourceAWSVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	pc, err := client.VPCPeeringConnections.GetVPCPeering(
		projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d, m))
	}

	return copyAWSVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceAWSVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())

	err := client.VPCPeeringConnections.DeleteVPCPeering(
		projectName,
		vpcID,
		peerCloudAccount,
		peerVPC,
		peerRegion,
	)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting VPC peering connection: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
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
				projectName,
				vpcID,
				peerCloudAccount,
				peerVPC,
				peerRegion,
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error waiting for AWS Aiven VPC Peering Connection to be DELETED: %s", err)
	}
	return nil
}

func resourceAWSVPCPeeringConnectionImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	m.(*meta.Meta).Import = true

	if len(strings.Split(d.Id(), "/")) != 5 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>/<aws_account_id>/<aws_vpc_id>/<aws_vpc_region>", d.Id())
	}

	dig := resourceAWSVPCPeeringConnectionRead(ctx, d, m)
	if dig.HasError() {
		return nil, errors.New("cannot get AWS VPC peering connection")
	}

	return []*schema.ResourceData{d}, nil
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
