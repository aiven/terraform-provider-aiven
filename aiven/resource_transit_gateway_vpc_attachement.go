// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenTransitGatewayVPCAttachmentSchema = map[string]*schema.Schema{
	"vpc_id": {
		Description: "The VPC the peering connection belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_cloud_account": {
		Description: "AWS account ID or GCP project ID of the peered VPC",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_vpc": {
		Description: "Transit gateway ID",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"user_peer_network_cidrs": {
		Description: "List of private IPv4 ranges to route through the peering connection",
		Required:    true,
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 128,
			MinItems: 1,
		},
	},
	"peer_region": {
		Description: "AWS region of the peered VPC (if not in the same region as Aiven VPC)",
		Required:    true,
		Type:        schema.TypeString,
	},
	"state": {
		Computed:    true,
		Description: "State of the peering connection",
		Type:        schema.TypeString,
	},
	"state_info": {
		Computed:    true,
		Description: "State-specific help or error information",
		Type:        schema.TypeMap,
	},
	"peering_connection_id": {
		Computed:    true,
		Description: "Cloud provider identifier for the peering connection if available",
		Type:        schema.TypeString,
	},
}

func resourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
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
	client := m.(*aiven.Client)

	cidrs := flattenToString(d.Get("user_peer_network_cidrs").([]interface{}))
	projectName, vpcID, peerCloudAccount, peerVPC, _ := parsePeeringVPCId(d.Id())

	peeringConnection, err := client.VPCPeeringConnections.Get(projectName, vpcID, peerCloudAccount, peerVPC)
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
			add = append(add, aiven.TransitGatewayVPCAttachment{
				CIDR:              fresh,
				PeerCloudAccount:  peerCloudAccount,
				PeerResourceGroup: peeringConnection.PeerResourceGroup,
				PeerVPC:           peerVPC,
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

	_, err = client.TransitGatewayVPCAttachment.Update(projectName, vpcID, aiven.TransitGatewayVPCAttachmentRequest{
		Add:    add,
		Delete: deleteCIDRs,
	})
	if err != nil {
		return diag.Errorf("cannot update transit gateway vpc attachment %s", err)
	}

	return resourceVPCPeeringConnectionRead(ctx, d, m)
}
