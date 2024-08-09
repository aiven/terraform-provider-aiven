package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAzureVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzureVPCPeeringConnectionRead,
		Description: "Gets information about about an Azure VPC peering connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAzureVPCPeeringConnectionSchema,
			"vpc_id", "azure_subscription_id", "peer_resource_group", "vnet_name", "peer_azure_app_id", "peer_azure_tenant_id"),
	}
}

func datasourceAzureVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subscriptionID := d.Get("azure_subscription_id").(string)
	vnetName := d.Get("vnet_name").(string)
	appID := d.Get("peer_azure_app_id").(string)
	tenantID := d.Get("peer_azure_tenant_id").(string)

	vpc, err := client.VPCs.Get(ctx, projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting Azure VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == subscriptionID && peer.PeerVPC == vnetName && peer.PeerAzureAppId == appID && peer.PeerAzureTenantId == tenantID {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			return resourceAzureVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("Azure peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, subscriptionID, vnetName)
}
