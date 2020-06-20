// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceVPCPeeringConnectionRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenVPCPeeringConnectionSchema, "vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}

func datasourceVPCPeeringConnectionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	peerCloudAccount := d.Get("peer_cloud_account").(string)
	peerVPC := d.Get("peer_vpc").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return err
	}
	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == peerCloudAccount && peer.PeerVPC == peerVPC {
			if peer.PeerRegion != nil && *peer.PeerRegion != "" {
				d.SetId(buildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC, *peer.PeerRegion))
			} else {
				d.SetId(buildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			}
			return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, peer, projectName, vpcID)
		}
	}

	return fmt.Errorf("Peering connection from %s/%s to %s/%s not found", projectName, vpc.CloudName, peerCloudAccount, peerVPC)
}
