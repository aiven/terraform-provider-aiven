package vpc

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/meta"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceAWSVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAWSVPCPeeringConnectionRead,
		Description: "The AWS VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAWSVPCPeeringConnectionSchema,
			"vpc_id", "aws_account_id", "aws_vpc_id", "aws_vpc_region"),
	}
}

func datasourceAWSVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, vpcID := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	awsAccountId := d.Get("aws_account_id").(string)
	awsVPCId := d.Get("aws_vpc_id").(string)
	awsVPCRegion := d.Get("aws_vpc_region").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting AWS VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == awsAccountId && peer.PeerVPC == awsVPCId && *peer.PeerRegion == awsVPCRegion {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC, awsVPCRegion))
			return resourceAWSVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("AWS peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, awsAccountId, awsVPCId)
}
