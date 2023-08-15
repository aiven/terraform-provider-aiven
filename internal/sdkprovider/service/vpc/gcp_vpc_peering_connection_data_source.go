package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
=======

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
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

	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	gcpProjectID := d.Get("gcp_project_id").(string)
	peerVPC := d.Get("peer_vpc").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting Azure VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == gcpProjectID && peer.PeerVPC == peerVPC {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			return resourceGCPVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("gcp peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, gcpProjectID, peerVPC)
}
