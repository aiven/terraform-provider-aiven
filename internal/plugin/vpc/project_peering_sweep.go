package vpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	for cloud, prefix := range map[string]string{"aws": "aws-", "azure": "azure-", "gcp": "google-"} {
		name := "aiven_" + cloud + "_vpc_peering_connection"
		sweep.AddTestSweepers(name, &resource.Sweeper{
			Name:         name,
			Dependencies: []string{"aiven_project_vpc"},
			F:            func(_ string) error { return sweepProjectPeerings(prefix) },
		})
	}
}

func sweepProjectPeerings(cloudPrefix string) error {
	ctx := context.Background()
	project := os.Getenv("AIVEN_PROJECT_NAME")
	client, err := sweep.SharedGenClient()
	if err != nil {
		return err
	}
	vpcs, err := client.VpcList(ctx, project)
	if err != nil {
		return err
	}
	for _, network := range vpcs {
		if !strings.HasPrefix(network.CloudName, cloudPrefix) {
			continue
		}
		networkDetails, err := client.VpcGet(ctx, project, network.ProjectVpcId)
		if adapter.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		for _, connection := range networkDetails.PeeringConnections {
			if connection.VpcPeeringConnectionType == vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment {
				continue
			}
			id := &ProjectPeeringID{
				Project: project, ProjectVPCID: network.ProjectVpcId,
				PeerCloudAccount: connection.PeerCloudAccount, PeerVPC: connection.PeerVpc, PeerRegion: connection.PeerRegion,
			}
			var group *string
			if cloudPrefix == "azure-" {
				if connection.PeerResourceGroup == "" {
					return fmt.Errorf("azure VPC peering %q has no resource group", id)
				}
				group = &connection.PeerResourceGroup
			}
			if err := id.Delete(ctx, client, group); err != nil && !adapter.IsNotFound(err) {
				return fmt.Errorf("deleting VPC peering %q: %w", id, err)
			}
		}
	}
	return nil
}
