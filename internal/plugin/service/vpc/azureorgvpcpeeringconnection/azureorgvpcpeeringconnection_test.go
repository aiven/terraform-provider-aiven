package azureorgvpcpeeringconnection

import (
	"errors"
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
)

func TestDeleteView(t *testing.T) {
	t.Run("deletes by API ID from state", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)
		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
			Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).
			Once()
		require.NoError(t, deleteView(t.Context(), client, d))
	})

	t.Run("requires API ID in state", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)
		require.NoError(t, d.Set("peering_connection_id", nil))

		require.EqualError(t, deleteView(t.Context(), client, d), "Azure organization VPC peering connection state has no API peering connection ID")
	})

	t.Run("propagates client error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)
		deleteErr := errors.New("delete failed")

		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
			Return(nil, deleteErr).
			Once()

		require.ErrorIs(t, deleteView(t.Context(), client, d), deleteErr)
	})
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
