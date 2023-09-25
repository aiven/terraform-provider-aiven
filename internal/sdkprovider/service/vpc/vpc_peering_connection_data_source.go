package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func datasourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	peerCloudAccount := d.Get("peer_cloud_account").(string)
	peerVPC := d.Get("peer_vpc").(string)

	vpc, err := client.VPCs.Get(ctx, projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == peerCloudAccount && peer.PeerVPC == peerVPC {
			if peer.PeerRegion != nil && *peer.PeerRegion != "" {
				d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC, *peer.PeerRegion))
			} else {
				d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			}
			return resourceVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, peerCloudAccount, peerVPC)
}
