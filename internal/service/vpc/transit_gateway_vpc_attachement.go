package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenTransitGatewayVPCAttachmentSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The VPC the peering connection belongs to.").ForceNew().Referenced().Build(),
	},
	"peer_cloud_account": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("AWS account ID or GCP project ID of the peered VPC").ForceNew().Build(),
	},
	"peer_vpc": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("Transit gateway ID").ForceNew().Build(),
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
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenTransitGatewayVPCAttachmentSchema,
	}
}

func resourceTransitGatewayVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing peering VPC ID: %s", err)
	}

	cidrs := schemautil.FlattenToString(d.Get("user_peer_network_cidrs").([]interface{}))
	peeringConnection, err := client.VPCPeeringConnections.Get(p.projectName, p.vpcID, p.peerCloudAccount, p.peerVPC)
	if err != nil {
		return diag.Errorf("cannot get transit gateway vpc attachment by id %s: %s", d.Id(), err)
	}

	// prepare a list of new transit gateway vpc attachment that needs to be added
	var add []aiven.TransitGatewayVPCAttachment
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
			if peeringConnection.PeerResourceGroup != nil {
				peerResourceGroup = peeringConnection.PeerResourceGroup
			}
			add = append(add, aiven.TransitGatewayVPCAttachment{
				CIDR:              fresh,
				PeerCloudAccount:  p.peerCloudAccount,
				PeerResourceGroup: peerResourceGroup,
				PeerVPC:           p.peerVPC,
			})
		}
	}

	// prepare a list of old cirds for deletion
	var deleteCIDRs []string
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

	_, err = client.TransitGatewayVPCAttachment.Update(p.projectName, p.vpcID, aiven.TransitGatewayVPCAttachmentRequest{
		Add:    add,
		Delete: deleteCIDRs,
	})
	if err != nil {
		return diag.Errorf("cannot update transit gateway vpc attachment %s", err)
	}

	return resourceVPCPeeringConnectionRead(ctx, d, m)
}
