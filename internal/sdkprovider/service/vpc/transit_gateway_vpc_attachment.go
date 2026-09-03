package vpc

import (
	"context"
	"slices"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
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
		Description: userconfig.Desc("AWS account ID, Google Cloud project ID, Azure subscription ID of the peered VPC, or string \"upcloud\" for UpCloud peering connections").ForceNew().Build(),
	},
	"peer_vpc": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("Peer network identifier. For AWS, the Transit Gateway ID; for Google Cloud, the VPC network name; for Azure, the resource group name; or for UpCloud, the network UUID)").ForceNew().Build(),
	},
	"user_peer_network_cidrs": {
		Required:    true,
		Type:        schema.TypeSet,
		Description: "List of private IPv4 ranges to route through the peering connection",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 128,
			MinItems: 1,
		},
	},
	"peer_region": {
		Optional:    true,
		Type:        schema.TypeString,
		ForceNew:    true,
		Description: "Region of the peered cloud provider resource, if not in the same region as Aiven VPC. This value can't be changed.",
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
		Description:   "Attach an Aiven virtual private cloud (VPC) to external networks from cloud providers. Supported peer cloud accounts include AWS, Google Cloud, Azure, and UpCloud. The Aiven documentation has more information on [VPC peering in Aiven](https://aiven.io/docs/platform/howto/list-vpc-peering).",
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

func resourceTransitGatewayVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing peering VPC ID: %s", err)
	}

	var cidrs []string

	cidrsv, ok := d.GetOk("user_peer_network_cidrs")
	if ok {
		cidrsva, ok := cidrsv.(*schema.Set)
		if ok {
			cidrs = schemautil.FlattenToString(cidrsva.List())
		}
	}

	peeringConnection, err := client.VPCPeeringConnections.Get(ctx, p.projectName, p.vpcID, p.peerCloudAccount, p.peerVPC)
	if err != nil {
		return diag.Errorf("cannot get transit gateway vpc attachment by id %s: %s", d.Id(), err)
	}

	// prepare a list of new transit gateway vpc attachment that needs to be added
	add := make([]aiven.TransitGatewayVPCAttachment, 0)
	for _, fresh := range cidrs {
		isNew := !slices.Contains(peeringConnection.UserPeerNetworkCIDRs, fresh)

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
	deleteCIDRs := make([]string, 0)
	for _, old := range peeringConnection.UserPeerNetworkCIDRs {
		forDeletion := true

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

	_, err = client.TransitGatewayVPCAttachment.Update(ctx, p.projectName, p.vpcID, aiven.TransitGatewayVPCAttachmentRequest{
		Add:    add,
		Delete: deleteCIDRs,
	})
	if err != nil {
		return diag.Errorf("cannot update transit gateway vpc attachment %s", err)
	}

	return resourceVPCPeeringConnectionRead(ctx, d, m)
}
