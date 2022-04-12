package vpc

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceAzureVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzureVPCPeeringConnectionRead,
		Description: "The Azure VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAzureVPCPeeringConnectionSchema,
			"vpc_id", "azure_subscription_id", "peer_resource_group", "vnet_name", "peer_azure_app_id", "peer_azure_tenant_id"),
	}
}

func datasourceAzureVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	subscriptionId := d.Get("azure_subscription_id").(string)
	vnetName := d.Get("vnet_name").(string)
	appId := d.Get("peer_azure_app_id").(string)
	tenantId := d.Get("peer_azure_tenant_id").(string)

	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.Errorf("Error getting Azure VPC peering connection: %s", err)
	}

	for _, peer := range vpc.PeeringConnections {
		if peer.PeerCloudAccount == subscriptionId && peer.PeerVPC == vnetName && peer.PeerAzureAppId == appId && peer.PeerAzureTenantId == tenantId {
			d.SetId(schemautil.BuildResourceID(projectName, vpcID, peer.PeerCloudAccount, peer.PeerVPC))
			return resourceAzureVPCPeeringConnectionRead(ctx, d, m)
		}
	}

	return diag.Errorf("Azure peering connection %s/%s/%s/%s not found",
		projectName, vpc.CloudName, subscriptionId, vnetName)
}
