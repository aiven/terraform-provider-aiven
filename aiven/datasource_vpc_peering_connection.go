// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVPCPeeringConnectionRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenVPCPeeringConnectionSchema,
			"vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}

func datasourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	peerCloudAccount := d.Get("peer_cloud_account").(string)
	peerVPC := d.Get("peer_vpc").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error deleting VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == peerCloudAccount && peer.PeerVPC == peerVPC {
			if peer.PeerRegion != nil && *peer.PeerRegion != "" {
				d.SetId(buildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC, *peer.PeerRegion))
			} else {
				d.SetId(buildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			}
			return resourceVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("peering connection from %s/%s to %s/%s not found",
		projectName, vpc.CloudName, peerCloudAccount, peerVPC)
}
