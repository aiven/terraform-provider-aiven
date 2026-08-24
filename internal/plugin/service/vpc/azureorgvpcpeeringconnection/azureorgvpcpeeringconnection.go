package azureorgvpcpeeringconnection

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

func flattenModifier(ctx context.Context, _ avngen.Client) adapter.MapModifier {
	return pluginvpc.PendingPeerWarning(ctx, "Azure")
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	organizationID := d.Get("organization_id").(string)
	organizationVPCID := d.Get("organization_vpc_id").(string)
	azureSubscriptionID := d.Get("azure_subscription_id").(string)
	vnetName := d.Get("vnet_name").(string)
	resourceGroup := d.Get("peer_resource_group").(string)

	return pluginvpc.DeleteOrgVPCPeeringConnection(
		ctx,
		client,
		organizationID,
		organizationVPCID,
		"Azure",
		func(connection *organizationvpc.OrganizationVpcGetPeeringConnectionOut) bool {
			return connection.PeerCloudAccount == azureSubscriptionID &&
				connection.PeerVpc == vnetName &&
				connection.PeerResourceGroup == resourceGroup
		},
	)
}
