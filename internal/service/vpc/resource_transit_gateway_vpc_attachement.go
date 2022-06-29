package vpc

import (
	"context"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenTransitGatewayVPCAttachmentSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("The VPC the peering connection belongs to.").ForceNew().Referenced().Build(),
	},
	"peer_cloud_account": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("AWS account ID or GCP project ID of the peered VPC").ForceNew().Build(),
	},
	"peer_vpc": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Transit gateway ID").ForceNew().Build(),
	},
	"user_peer_network_cidrs": {
		Required:    true,
		Type:        schema.TypeList,
		Description: "List of private IPv4 ranges to route through the peering connection",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 128,
			MinItems: 1,
		},
	},
	"peer_region": {
		Required:    true,
		Type:        schema.TypeString,
		Description: "AWS region of the peered VPC (if not in the same region as Aiven VPC)",
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
	"peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "Cloud provider identifier for the peering connection if available",
	},
}

func ResourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		Description:   "The Transit Gateway VPC Attachment resource allows the creation and management Transit Gateway VPC Attachment VPC peering connection between Aiven and AWS.",
		CreateContext: resourceVPCPeeringConnectionCreate,
		ReadContext:   resourceVPCPeeringConnectionRead,
		UpdateContext: resourceTransitGatewayVPCAttachmentUpdate,
		DeleteContext: resourceVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPeeringConnectionImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenTransitGatewayVPCAttachmentSchema,
	}
}

func resourceTransitGatewayVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	cidrs := schemautil.FlattenToString(d.Get("user_peer_network_cidrs").([]interface{}))
	projectName, vpcID, peerCloudAccount, peerVPC, _ := parsePeeringVPCId(d.Id())

	peeringConnection, err := client.VPCPeeringConnections.Get(projectName, vpcID, peerCloudAccount, peerVPC)
	if err != nil {
		return diag.Errorf("cannot get transit gateway vpc attachment by id %s: %s", d.Id(), err)
	}

	// prepare a list of new transit gateway vpc attachment that needs to be added
	add := []aiven.TransitGatewayVPCAttachment{}
	for _, fresh := range cidrs {
		var isNew = true

		for _, old := range peeringConnection.UserPeerNetworkCIDRs {
			if fresh == old {
				isNew = false
				break
			}
		}

		if isNew {
			var peerResourceGroup *string
			if len(peeringConnection.PeerResourceGroup) > 0 {
				peerResourceGroup = aiven.ToStringPointer(peeringConnection.PeerResourceGroup)
			}
			add = append(add, aiven.TransitGatewayVPCAttachment{
				CIDR:              fresh,
				PeerCloudAccount:  peerCloudAccount,
				PeerResourceGroup: peerResourceGroup,
				PeerVPC:           peerVPC,
			})
		}
	}

	// prepare a list of old cirds for deletion
	deleteCIDRs := []string{}
	for _, old := range peeringConnection.UserPeerNetworkCIDRs {
		var forDeletion = true

		for _, fresh := range cidrs {
			if old == fresh {
				forDeletion = false
			}
		}

		if forDeletion {
			deleteCIDRs = append(deleteCIDRs, old)
		}
	}

	if len(add) == 0 && len(deleteCIDRs) == 0 {
		return resourceVPCPeeringConnectionRead(ctx, d, m)
	}

	_, err = client.TransitGatewayVPCAttachment.Update(projectName, vpcID, aiven.TransitGatewayVPCAttachmentRequest{
		Add:    add,
		Delete: deleteCIDRs,
	})
	if err != nil {
		return diag.Errorf("cannot update transit gateway vpc attachment %s", err)
	}

	return resourceVPCPeeringConnectionRead(ctx, d, m)
}
