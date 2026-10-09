package awsorgvpcpeeringconnection

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

func flattenModifier(ctx context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		func(d adapter.ResourceData, dto map[string]any) error {
			if err := d.Set("aws_vpc_peering_connection_id", nil); err != nil {
				return err
			}
			if stateInfo, ok := dto["state_info"].(map[string]any); ok {
				if connectionID, ok := stateInfo["aws_vpc_peering_connection_id"].(string); ok {
					dto["aws_vpc_peering_connection_id"] = connectionID
				}
			}

			return nil
		},
		pluginvpc.PendingPeerWarning(ctx, "AWS"),
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	organizationID := d.Get("organization_id").(string)
	organizationVPCID := d.Get("organization_vpc_id").(string)
	awsAccountID := d.Get("aws_account_id").(string)
	awsVPCID := d.Get("aws_vpc_id").(string)
	awsRegion := d.Get("aws_vpc_region").(string)

	return pluginvpc.DeleteOrgVPCPeeringConnection(
		ctx,
		client,
		organizationID,
		organizationVPCID,
		"AWS",
		func(connection *organizationvpc.OrganizationVpcGetPeeringConnectionOut) bool {
			return connection.PeerCloudAccount == awsAccountID &&
				connection.PeerVpc == awsVPCID &&
				connection.PeerRegion != nil &&
				*connection.PeerRegion == awsRegion
		},
	)
}
