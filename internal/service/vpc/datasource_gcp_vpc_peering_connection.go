package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceGCPVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGCPVPCPeeringConnectionRead,
		Description: "The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenGCPVPCPeeringConnectionSchema,
			"vpc_id", "gcp_project_id", "peer_vpc"),
	}
}

func datasourceGCPVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	gcpProjectId := d.Get("gcp_project_id").(string)
	peerVPC := d.Get("peer_vpc").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting Azure VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == gcpProjectId && peer.PeerVPC == peerVPC {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			return resourceGCPVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("gcp peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, gcpProjectId, peerVPC)
}
