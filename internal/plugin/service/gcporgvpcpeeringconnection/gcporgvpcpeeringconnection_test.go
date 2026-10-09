package gcporgvpcpeeringconnection

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testOrganizationID    = "example-organization-id"
	testOrganizationVPCID = "example-organization-vpc-id"
	testGCPProjectID      = "example-gcp-project"
	testPeerVPC           = "example-peer-vpc"
	testConnectionID      = "example-connection-id"
)

func TestFlattenModifierBuildsSelfLink(t *testing.T) {
	d := newTestResourceData(t)
	dto := map[string]any{
		"state": string(organizationvpc.VpcPeeringConnectionStateTypeActive),
		"state_info": map[string]any{
			"to_project_id":  "peer-project",
			"to_vpc_network": "peer-network",
		},
	}

	require.NoError(t, flattenModifier(t.Context(), nil)(d, dto))
	require.Equal(t, gcpAPI+"/projects/peer-project/global/networks/peer-network", dto["self_link"])
}

func TestDeleteView(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newTestResourceData(t)

	client.EXPECT().
		OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
		Return(&organizationvpc.OrganizationVpcGetOut{
			PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
				{
					PeerCloudAccount:    "another-gcp-project",
					PeerVpc:             testPeerVPC,
					PeeringConnectionId: new("another-project-connection-id"),
				},
				{
					PeerCloudAccount:    testGCPProjectID,
					PeerVpc:             "another-peer-vpc",
					PeeringConnectionId: new("another-vpc-connection-id"),
				},
				{
					PeerCloudAccount:    testGCPProjectID,
					PeerVpc:             testPeerVPC,
					PeeringConnectionId: new(testConnectionID),
				},
			},
		}, nil).
		Once()
	client.EXPECT().
		OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
		Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).
		Once()

	require.NoError(t, deleteView(t.Context(), client, d))
}

func newTestResourceData(t *testing.T) adapter.ResourceData {
	t.Helper()

	values := map[string]any{
		"organization_id":     testOrganizationID,
		"organization_vpc_id": testOrganizationVPCID,
		"gcp_project_id":      testGCPProjectID,
		"peer_vpc":            testPeerVPC,
		"id":                  testOrganizationID + "/" + testOrganizationVPCID + "/" + testGCPProjectID + "/" + testPeerVPC,
	}

	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(values))
	require.NoError(t, err)
	return d
}
