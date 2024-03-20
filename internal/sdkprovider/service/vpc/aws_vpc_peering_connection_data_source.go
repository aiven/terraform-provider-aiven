package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAWSVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAWSVPCPeeringConnectionRead,
		Description: "Gets information about an AWS VPC peering connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAWSVPCPeeringConnectionSchema,
			"vpc_id", "aws_account_id", "aws_vpc_id", "aws_vpc_region"),
	}
}

func datasourceAWSVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	awsAccountID := d.Get("aws_account_id").(string)
	awsVPCID := d.Get("aws_vpc_id").(string)
	awsVPCRegion := d.Get("aws_vpc_region").(string)

	vpc, err := client.VPCs.Get(ctx, projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting AWS VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == awsAccountID && peer.PeerVPC == awsVPCID && *peer.PeerRegion == awsVPCRegion {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC, awsVPCRegion))
			return resourceAWSVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("AWS peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, awsAccountID, awsVPCID)
}
