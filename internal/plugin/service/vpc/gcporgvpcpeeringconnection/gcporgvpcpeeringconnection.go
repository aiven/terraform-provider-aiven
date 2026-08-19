package gcporgvpcpeeringconnection

import (
	"context"
	"errors"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

const gcpAPI = "https://www.googleapis.com/compute/v1"

func flattenModifier(ctx context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		func(_ adapter.ResourceData, dto map[string]any) error {
			stateInfo, ok := dto["state_info"].(map[string]any)
			if ok {
				projectID, projectOK := stateInfo["to_project_id"].(string)
				vpcNetwork, networkOK := stateInfo["to_vpc_network"].(string)
				if projectOK && networkOK {
					dto["self_link"] = fmt.Sprintf("%s/projects/%s/global/networks/%s", gcpAPI, projectID, vpcNetwork)
				}
			}

			return nil
		},
		pluginvpc.PendingPeerWarning(ctx, "Google Cloud"),
	)
}

// deleteView resolves the API-only peering connection ID before deleting.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	organizationID := d.Get("organization_id").(string)
	organizationVPCID := d.Get("organization_vpc_id").(string)

	vpc, err := client.OrganizationVpcGet(ctx, organizationID, organizationVPCID)
	if adapter.IsNotFound(err) {
		return err
	}
	if err != nil {
		return fmt.Errorf("getting organization VPC %q: %w", organizationVPCID, err)
	}

	var connection *organizationvpc.OrganizationVpcGetPeeringConnectionOut
	gcpProjectID := d.Get("gcp_project_id").(string)
	peerVPC := d.Get("peer_vpc").(string)
	for i := range vpc.PeeringConnections {
		candidate := &vpc.PeeringConnections[i]
		if candidate.PeerCloudAccount == gcpProjectID &&
			candidate.PeerVpc == peerVPC {
			connection = candidate
			break
		}
	}
	if connection == nil {
		return adapter.ErrNotFound
	}
	if connection.PeeringConnectionId == nil {
		return errors.New("GCP organization VPC peering connection API response has no peering connection ID")
	}

	_, err = client.OrganizationVpcPeeringConnectionDeleteById(
		ctx,
		organizationID,
		organizationVPCID,
		*connection.PeeringConnectionId,
	)
	return err
}
