package gcporgvpcpeeringconnection

import (
	"context"
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
	gcpProjectID := d.Get("gcp_project_id").(string)
	peerVPC := d.Get("peer_vpc").(string)

	return pluginvpc.DeleteOrgVPCPeeringConnection(
		ctx,
		client,
		organizationID,
		organizationVPCID,
		"GCP",
		func(connection *organizationvpc.OrganizationVpcGetPeeringConnectionOut) bool {
			return connection.PeerCloudAccount == gcpProjectID &&
				connection.PeerVpc == peerVPC
		},
	)
}
