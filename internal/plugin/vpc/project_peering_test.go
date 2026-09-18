package vpc

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestProjectPeeringID(t *testing.T) {
	for _, value := range []string{"project/vpc/account/network", "project/vpc/account/network/region", "project/vpc/account/network/"} {
		id, err := ParseProjectPeeringID(value)
		require.NoError(t, err)
		require.Equal(t, value, id.String())
	}
	for _, value := range []string{"project/vpc/network", "project/vpc/account/network/region/extra"} {
		_, err := ParseProjectPeeringID(value)
		require.Error(t, err)
	}
}

func TestProjectPeeringFind(t *testing.T) {
	for _, tc := range []struct {
		id      string
		regions []*string
		want    *string
		wantErr error
	}{
		{id: "project/vpc/account/network", regions: []*string{new("region")}, want: new("region")},
		{id: "project/vpc/account/network", regions: []*string{new("region"), new("other")}, wantErr: adapter.ErrMultiple},
		{id: "project/vpc/account/network/region", regions: []*string{new("other"), new("region")}, want: new("region")},
		{id: "project/vpc/account/network/", regions: []*string{nil}},
		{id: "project/vpc/account/network/", regions: []*string{new("")}, want: new("")},
		{id: "project/vpc/account/network/missing", regions: []*string{new("region")}, wantErr: adapter.ErrNotFound},
	} {
		client := avngen.NewMockClient(t)
		connections := []vpc.PeeringConnectionOut{
			{PeerCloudAccount: "other-account", PeerVpc: "network"},
			{PeerCloudAccount: "account", PeerVpc: "other-network"},
		}
		for _, region := range tc.regions {
			connections = append(connections, vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "network", PeerRegion: region})
		}
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: connections}, nil).Once()
		id, err := ParseProjectPeeringID(tc.id)
		require.NoError(t, err)
		connection, err := id.Find(t.Context(), client)
		if tc.wantErr != nil {
			require.ErrorIs(t, err, tc.wantErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.want, connection.PeerRegion)
		}
	}
}

func TestProjectPeeringDeleteWithRegion(t *testing.T) {
	client := avngen.NewMockClient(t)
	id, err := ParseProjectPeeringID("project/vpc/account/network/region")
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "account", "network", "region").
		Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	require.NoError(t, id.Delete(t.Context(), client, nil))
}

func TestStateInfoMap(t *testing.T) {
	f := func(info map[string]any, want map[string]string) {
		t.Helper()

		require.Equal(t, want, StateInfoMap(info))
	}

	f(nil, nil)
	f(map[string]any{}, nil)
	f(map[string]any{
		"aws_transit_gateway_attachment_id": "tgw-attach-123",
		"null":                              nil,
		"empty":                             "",
		"future_nested":                     map[string]any{"details": []any{"alpha", 42}},
		"message":                           "attachment available",
		"type":                              "attachment-ready",
	}, map[string]string{
		"null":                              "<nil>",
		"aws_transit_gateway_attachment_id": "tgw-attach-123",
		"empty":                             "",
		"future_nested":                     "map[details:[alpha 42]]",
		"message":                           "attachment available",
		"type":                              "attachment-ready",
	})
}

func TestPeeringRefreshStateCheck(t *testing.T) {
	for _, tc := range []struct {
		state  string
		failed bool
		done   bool
		detail string
	}{
		{state: "ACTIVE", done: true},
		{state: "PENDING_PEER", done: true},
		{state: "APPROVED", detail: "transient state"},
		{state: "APPROVED_PEER_REQUESTED", detail: "transient state"},
		{state: "NEW_BACKEND_STATE", detail: `unknown VPC peering connection state "NEW_BACKEND_STATE"`},
		{state: "REJECTED_BY_PEER", failed: true, detail: "rejected by the peer"},
		{state: "INVALID_SPECIFICATION", failed: true, detail: "specification is invalid"},
		{state: "DELETED", failed: true, detail: "was deleted"},
		{state: "DELETING", failed: true, detail: "was deleted"},
		{state: "DELETED_BY_PEER", failed: true, detail: "peer cloud resource was deleted"},
		{state: "ERROR", failed: true, detail: "reached ERROR"},
	} {
		for _, info := range []map[string]any{nil, {"message": "cloud detail", "type": "cloud-state"}} {
			d, err := adapter.NewResourceData(&adapter.Schema{Type: adapter.SchemaTypeObject, Properties: map[string]*adapter.Schema{
				"id": {Type: adapter.SchemaTypeString}, "state": {Type: adapter.SchemaTypeString},
				"state_info": {Type: adapter.SchemaTypeMap, Items: &adapter.Schema{Type: adapter.SchemaTypeString}},
			}}, []string{"id"}, adapter.WithTestState(map[string]any{"state": tc.state, "state_info": info}))
			require.NoError(t, err)
			err = PeeringRefreshStateCheck(d)
			switch {
			case tc.done:
				require.NoError(t, err)
			case tc.failed:
				require.ErrorIs(t, err, adapter.ErrRefreshStateFailed)
			default:
				require.Error(t, err)
				require.NotErrorIs(t, err, adapter.ErrRefreshStateFailed)
			}
			if !tc.done {
				require.ErrorContains(t, err, tc.detail)
				if info != nil {
					require.ErrorContains(t, err, `state_info: message="cloud detail", type="cloud-state"`)
				}
			}
		}
	}
}
