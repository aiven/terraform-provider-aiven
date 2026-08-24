package vpc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// DeleteOrgVPCPeeringConnection resolves the current API ID and deletes the first matching connection.
func DeleteOrgVPCPeeringConnection(
	ctx context.Context,
	client avngen.Client,
	organizationID string,
	organizationVPCID string,
	cloud string,
	match func(*organizationvpc.OrganizationVpcGetPeeringConnectionOut) bool,
) error {
	vpc, err := client.OrganizationVpcGet(ctx, organizationID, organizationVPCID)
	if adapter.IsNotFound(err) {
		return err
	}
	if err != nil {
		return fmt.Errorf("getting organization VPC %q: %w", organizationVPCID, err)
	}

	for i := range vpc.PeeringConnections {
		connection := &vpc.PeeringConnections[i]
		if !match(connection) {
			continue
		}
		if connection.PeeringConnectionId == nil {
			return fmt.Errorf("%s organization VPC peering connection API response has no peering connection ID", cloud)
		}

		_, err = client.OrganizationVpcPeeringConnectionDeleteById(
			ctx,
			organizationID,
			organizationVPCID,
			*connection.PeeringConnectionId,
		)
		return err
	}

	return adapter.ErrNotFound
}
