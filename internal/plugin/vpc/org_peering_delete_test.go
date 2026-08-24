package vpc

import (
	"errors"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testOrganizationID        = "example-organization-id"
	testOrganizationVPCID     = "example-organization-vpc-id"
	testPeerCloudAccount      = "example-peer-cloud-account"
	testCurrentConnectionID   = "current-connection-id"
	testAnotherConnectionID   = "another-connection-id"
	testUnrelatedConnectionID = "unrelated-connection-id"
)

func TestDeleteOrgVPCPeeringConnection(t *testing.T) {
	match := func(connection *organizationvpc.OrganizationVpcGetPeeringConnectionOut) bool {
		return connection.PeerCloudAccount == testPeerCloudAccount
	}

	t.Run("deletes the first matching connection by its current API ID", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(&organizationvpc.OrganizationVpcGetOut{
				PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
					{
						PeerCloudAccount:    "another-peer-cloud-account",
						PeeringConnectionId: new(testUnrelatedConnectionID),
					},
					{
						PeerCloudAccount:    testPeerCloudAccount,
						PeeringConnectionId: new(testCurrentConnectionID),
					},
					{
						PeerCloudAccount:    testPeerCloudAccount,
						PeeringConnectionId: new(testAnotherConnectionID),
					},
				},
			}, nil).
			Once()
		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testCurrentConnectionID).
			Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).
			Once()

		require.NoError(t, DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		))
	})

	t.Run("returns missing parent", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(nil, avngen.Error{Status: http.StatusNotFound}).
			Once()

		err := DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		)
		require.True(t, adapter.IsNotFound(err))
	})

	t.Run("propagates get error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		getErr := errors.New("get failed")
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(nil, getErr).
			Once()

		err := DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		)
		require.ErrorIs(t, err, getErr)
	})

	t.Run("returns missing connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(&organizationvpc.OrganizationVpcGetOut{}, nil).
			Once()

		err := DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		)
		require.ErrorIs(t, err, adapter.ErrNotFound)
	})

	t.Run("rejects the first matching connection without an API ID", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(&organizationvpc.OrganizationVpcGetOut{
				PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
					{PeerCloudAccount: testPeerCloudAccount},
					{
						PeerCloudAccount:    testPeerCloudAccount,
						PeeringConnectionId: new(testAnotherConnectionID),
					},
				},
			}, nil).
			Once()

		err := DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		)
		require.EqualError(t, err, "AWS organization VPC peering connection API response has no peering connection ID")
		require.NotErrorIs(t, err, adapter.ErrNotFound)
	})

	t.Run("propagates delete error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		deleteErr := errors.New("delete failed")
		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(&organizationvpc.OrganizationVpcGetOut{
				PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{{
					PeerCloudAccount:    testPeerCloudAccount,
					PeeringConnectionId: new(testCurrentConnectionID),
				}},
			}, nil).
			Once()
		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testCurrentConnectionID).
			Return(nil, deleteErr).
			Once()

		err := DeleteOrgVPCPeeringConnection(
			t.Context(),
			client,
			testOrganizationID,
			testOrganizationVPCID,
			"AWS",
			match,
		)
		require.ErrorIs(t, err, deleteErr)
	})
}
