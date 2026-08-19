package azureorgvpcpeeringconnection

import (
	"context"
	"errors"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

func flattenModifier(ctx context.Context, _ avngen.Client) adapter.MapModifier {
	return pluginvpc.PendingPeerWarning(ctx, "Azure")
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	connectionID := d.Get("peering_connection_id").(string)
	if connectionID == "" {
		return errors.New("Azure organization VPC peering connection state has no API peering connection ID") // nolint:staticcheck
	}

	_, err := client.OrganizationVpcPeeringConnectionDeleteById(
		ctx,
		d.Get("organization_id").(string),
		d.Get("organization_vpc_id").(string),
		connectionID,
	)
	return err
}
