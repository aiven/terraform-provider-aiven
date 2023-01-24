package vpc

import (
	"context"
	"fmt"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const _gcpAPI = "https://www.googleapis.com/compute/v1"

var aivenGCPVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		Description:  userconfig.Desc("The VPC the peering connection belongs to.").ForceNew().Build(),
		ValidateFunc: validateVPCID,
	},
	"gcp_project_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("GCP project ID.").ForceNew().Build(),
	},
	"peer_vpc": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("GCP VPC network name.").ForceNew().Build(),
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
	"self_link": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "Computed GCP network peering link",
	},
}

func ResourceGCPVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "The GCP VPC Peering Connection resource allows the creation and management of Aiven GCP VPC Peering Connections.",
		CreateContext: resourceGCPVPCPeeringConnectionCreate,
		ReadContext:   resourceGCPVPCPeeringConnectionRead,
		DeleteContext: resourceGCPVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenGCPVPCPeeringConnectionSchema,
	}
}

func resourceGCPVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc  *aiven.VPCPeeringConnection
		err error
	)

	client := m.(*aiven.Client)
	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	gcpProjectID := d.Get("gcp_project_id").(string)
	peerVPC := d.Get("peer_vpc").(string)

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		projectName, vpcID, gcpProjectID, peerVPC, nil)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("error checking gcp peering connection: %s", err)
	}

	if pc != nil {
		return diag.Errorf("gcp vpc peering connection already exists and cannot be created")
	}

	if _, err = client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount: gcpProjectID,
			PeerVPC:          peerVPC,
		},
	); err != nil {
		return diag.Errorf("Error waiting for VPC peering connection creation: %s", err)
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
				gcpProjectID,
				peerVPC,
				nil,
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

	d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		return append(diags, resourceGCPVPCPeeringConnectionDelete(ctx, d, m)...)
	}

	return append(diags, resourceGCPVPCPeeringConnectionRead(ctx, d, m)...)
}

func resourceGCPVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing GCP peering VPC ID: %s", err)
	}

	var pc *aiven.VPCPeeringConnection
	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		p.projectName, p.vpcID, p.peerCloudAccount, p.peerVPC, p.peerRegion)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	return copyGCPVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, p.projectName, p.vpcID)
}

func resourceGCPVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing GCP peering VPC ID: %s", err)
	}
	err = client.VPCPeeringConnections.DeleteVPCPeering(
		p.projectName,
		p.vpcID,
		p.peerCloudAccount,
		p.peerVPC,
		p.peerRegion,
	)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting GCP VPC peering connection: %s", err)
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
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error waiting for GCP Aiven VPC Peering Connection to be DELETED: %s", err)
	}
	return nil
}

func copyGCPVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) diag.Diagnostics {
	if err := d.Set("vpc_id", schemautil.BuildResourceID(project, vpcID)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gcp_project_id", peeringConnection.PeerCloudAccount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("peer_vpc", peeringConnection.PeerVPC); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state", peeringConnection.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state_info", ConvertStateInfoToMap(peeringConnection.StateInfo)); err != nil {
		return diag.FromErr(err)
	}

	var toProjectID string
	var toVPCNetwork string

	if peeringConnection.StateInfo != nil {
		si := *peeringConnection.StateInfo
		if len(si) > 0 {
			if v, ok := si["to_project_id"]; ok {
				toProjectID = v.(string)
			}
			if v, ok := si["to_vpc_network"]; ok {
				toVPCNetwork = v.(string)
			}

			if toProjectID != "" && toVPCNetwork != "" {
				if err := d.Set("self_link",
					fmt.Sprintf(_gcpAPI+"/projects/%s/global/networks/%s",
						toProjectID, toVPCNetwork)); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}
