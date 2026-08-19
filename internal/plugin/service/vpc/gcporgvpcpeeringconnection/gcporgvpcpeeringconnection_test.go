package gcporgvpcpeeringconnection

import (
	"net/http"
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
	orgVPC := func() *organizationvpc.OrganizationVpcGetOut {
		return &organizationvpc.OrganizationVpcGetOut{
			PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{{
				PeerCloudAccount:    testGCPProjectID,
				PeerVpc:             testPeerVPC,
				PeeringConnectionId: new(testConnectionID),
				State:               organizationvpc.VpcPeeringConnectionStateTypeActive,
				StateInfo: organizationvpc.PeeringConnectionStateInfoOut{
					Message:      "example state information",
					Type:         "example",
					ToProjectId:  new("peer-project"),
					ToVpcNetwork: new("peer-network"),
				},
			}},
		}
	}

	t.Run("finds API ID and deletes connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)

		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(orgVPC(), nil).
			Once()
		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
			Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).
			Once()
		require.NoError(t, deleteView(t.Context(), client, d))
	})

	t.Run("returns missing parent", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)

		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(nil, avngen.Error{Status: http.StatusNotFound}).
			Once()

		err := deleteView(t.Context(), client, d)
		require.True(t, adapter.IsNotFound(err))
	})

	t.Run("returns missing connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)

		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(&organizationvpc.OrganizationVpcGetOut{}, nil).
			Once()

		require.ErrorIs(t, deleteView(t.Context(), client, d), adapter.ErrNotFound)
	})

	t.Run("rejects connection without API ID", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)
		orgVPC := orgVPC()
		orgVPC.PeeringConnections[0].PeeringConnectionId = nil

		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(orgVPC, nil).
			Once()

		err := deleteView(t.Context(), client, d)
		require.EqualError(t, err, "GCP organization VPC peering connection API response has no peering connection ID")
		require.NotErrorIs(t, err, adapter.ErrNotFound)
	})

	t.Run("propagates delete error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newTestResourceData(t)

		client.EXPECT().
			OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
			Return(orgVPC(), nil).
			Once()
		client.EXPECT().
			OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
			Return(nil, avngen.Error{Status: http.StatusForbidden, Message: "forbidden"}).
			Once()

		err := deleteView(t.Context(), client, d)
		require.Error(t, err)
		var apiErr avngen.Error
		require.ErrorAs(t, err, &apiErr)
		require.Equal(t, http.StatusForbidden, apiErr.Status)
	})
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
