package vpc

import (
	"context"
	"errors"
	"fmt"
	"log"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	const sharedSweeper = "aiven_transit_gateway_vpc_attachment"
	sweep.AddTestSweepers(sharedSweeper, &resource.Sweeper{
		Name: sharedSweeper,
		F: func(_ string) error {
			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}
			return sweepProjectPeerings(context.Background(), client, sweep.ProjectName())
		},
	})
	// All four resources share the project peering API. Dependencies make the
	// runner perform one shared sweep even when several resource names are selected.
	for _, cloud := range []string{"aws", "azure", "gcp"} {
		name := "aiven_" + cloud + "_vpc_peering_connection"
		sweep.AddTestSweepers(name, &resource.Sweeper{
			Name:         name,
			Dependencies: []string{sharedSweeper},
			F:            func(_ string) error { return nil },
		})
	}
}

func sweepProjectPeerings(ctx context.Context, client avngen.Client, project string) error {
	vpcs, err := client.VpcList(ctx, project)
	if err != nil {
		return err
	}
	var result error
	for _, network := range vpcs {
		networkDetails, err := client.VpcGet(ctx, project, network.ProjectVpcId)
		if adapter.IsNotFound(err) {
			continue
		}
		if err != nil {
			result = errors.Join(result, fmt.Errorf("reading VPC %q: %w", network.ProjectVpcId, err))
			continue
		}
		for _, connection := range networkDetails.PeeringConnections {
			id := &ProjectPeeringID{
				Project: project, ProjectVPCID: network.ProjectVpcId,
				PeerCloudAccount: connection.PeerCloudAccount, PeerVPC: connection.PeerVpc,
			}
			if connection.PeerRegion != nil && *connection.PeerRegion != "" {
				id.PeerRegion = connection.PeerRegion
			}
			// These states reject DELETE but do not block deletion of the parent VPC.
			switch connection.State {
			case vpc.VpcPeeringConnectionStateTypeDeleting:
				continue
			case vpc.VpcPeeringConnectionStateTypeApprovedPeerRequested:
				log.Printf("[WARN] Skipping VPC peering %q: the API does not allow deletion in state %s; connection left in place", id, connection.State)
				continue
			}
			var group *string
			if connection.PeerResourceGroup != "" {
				group = &connection.PeerResourceGroup
			}
			if err := id.Delete(ctx, client, group); err != nil && !adapter.IsNotFound(err) {
				result = errors.Join(result, fmt.Errorf("deleting VPC peering %q: %w", id, err))
			}
		}
	}
	return result
}
