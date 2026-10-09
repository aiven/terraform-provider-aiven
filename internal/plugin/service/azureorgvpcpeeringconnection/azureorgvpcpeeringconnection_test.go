package azureorgvpcpeeringconnection

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testOrganizationID      = "example-organization-id"
	testOrganizationVPCID   = "example-organization-vpc-id"
	testAzureSubscriptionID = "00000000-0000-0000-0000-000000000001"
	testVNetName            = "example-vnet"
	testResourceGroup       = "example-resource-group"
	testConnectionID        = "example-connection-id"
	testStaleConnectionID   = "stale-connection-id"
)

func TestDeleteView(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newTestResourceData(t)
	require.NoError(t, d.Set("peering_connection_id", testStaleConnectionID))

	client.EXPECT().
		OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
		Return(&organizationvpc.OrganizationVpcGetOut{
			PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
				{
					PeerCloudAccount:    "another-subscription-id",
					PeerVpc:             testVNetName,
					PeerResourceGroup:   testResourceGroup,
					PeeringConnectionId: new("another-subscription-connection-id"),
				},
				{
					PeerCloudAccount:    testAzureSubscriptionID,
					PeerVpc:             "another-vnet",
					PeerResourceGroup:   testResourceGroup,
					PeeringConnectionId: new("another-vnet-connection-id"),
				},
				{
					PeerCloudAccount:    testAzureSubscriptionID,
					PeerVpc:             testVNetName,
					PeerResourceGroup:   "another-resource-group",
					PeeringConnectionId: new("another-resource-group-connection-id"),
				},
				{
					PeerCloudAccount:    testAzureSubscriptionID,
					PeerVpc:             testVNetName,
					PeerResourceGroup:   testResourceGroup,
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
		"organization_id":       testOrganizationID,
		"organization_vpc_id":   testOrganizationVPCID,
		"azure_subscription_id": testAzureSubscriptionID,
		"vnet_name":             testVNetName,
		"peer_resource_group":   testResourceGroup,
		"id":                    testOrganizationID + "/" + testOrganizationVPCID + "/" + testAzureSubscriptionID + "/" + testVNetName + "/" + testResourceGroup,
		"peering_connection_id": testConnectionID,
	}

	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(values))
	require.NoError(t, err)
	return d
}
