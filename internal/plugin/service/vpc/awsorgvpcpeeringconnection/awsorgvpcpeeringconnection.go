package awsorgvpcpeeringconnection

import (
	"context"
	"errors"

	avngen "github.com/aiven/go-client-codegen"

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
	connectionID := d.Get("peering_connection_id").(string)
	if connectionID == "" {
		return errors.New("AWS organization VPC peering connection state has no API peering connection ID")
	}

	_, err := client.OrganizationVpcPeeringConnectionDeleteById(
		ctx,
		d.Get("organization_id").(string),
		d.Get("organization_vpc_id").(string),
		connectionID,
	)
	return err
}
