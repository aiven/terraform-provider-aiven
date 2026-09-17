package gcpvpcpeeringconnection

import (
	"errors"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const testID = "aiven-project/aiven-vpc/gcp-project/peer-network"

func TestCreateView(t *testing.T) {
	t.Run("creates a regionless peering", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
			"vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
		}))
		require.NoError(t, err)
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "aiven-project", "aiven-vpc", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "gcp-project", PeerVpc: "peer-network",
		}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
		require.NoError(t, createView(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
	})

	t.Run("fails on create error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
			"vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
		}))
		require.NoError(t, err)
		createError := errors.New("create failed")
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "aiven-project", "aiven-vpc", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "gcp-project", PeerVpc: "peer-network",
		}).Return(&vpc.VpcPeeringConnectionCreateOut{}, createError).Once()
		require.ErrorIs(t, createView(t.Context(), client, d), createError)
		require.Empty(t, d.ID(), "failed Create must not adopt a peering")
	})
}

func TestReadViewLegacyStateAndImport(t *testing.T) {
	t.Run("import", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newReadData(t, map[string]any{"id": testID})
		connection := testConnection()
		connection.StateInfo = map[string]any{
			"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network",
			"attempts": 2, "ready": true, "empty": "", "cidrs": []any{"10.0.0.0/24"},
			"null": nil, "nested": map[string]any{"message": "detail"},
		}
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
		require.Equal(t, "aiven-project/aiven-vpc", d.Get("vpc_id"))
		require.Equal(t, "gcp-project", d.Get("gcp_project_id"))
		require.Equal(t, "peer-network", d.Get("peer_vpc"))
		require.Equal(t, "ACTIVE", d.Get("state"))
		require.Equal(t, "https://www.googleapis.com/compute/v1/projects/aiven-gcp-project/global/networks/aiven-network", d.Get("self_link"))
		require.Equal(t, map[string]any{
			"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network",
			"attempts": "2", "ready": "true", "empty": "", "cidrs": "[10.0.0.0/24]",
			"null": "<nil>", "nested": "map[message:detail]",
		}, d.Get("state_info"))
	})

	t.Run("SDK state", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newReadData(t, map[string]any{
			"id": testID, "vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
			"state": "PENDING_PEER", "state_info": map[string]any{"old": "value"}, "self_link": "old-link",
		})
		connection := testConnection()
		connection.StateInfo = map[string]any{
			"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network",
			"attempts": 2, "ready": true, "empty": "", "cidrs": []any{"10.0.0.0/24"},
			"null": nil, "nested": map[string]any{"message": "detail"},
		}
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
		require.Equal(t, "aiven-project/aiven-vpc", d.Get("vpc_id"))
		require.Equal(t, "gcp-project", d.Get("gcp_project_id"))
		require.Equal(t, "peer-network", d.Get("peer_vpc"))
		require.Equal(t, "ACTIVE", d.Get("state"))
		require.Equal(t, "https://www.googleapis.com/compute/v1/projects/aiven-gcp-project/global/networks/aiven-network", d.Get("self_link"))
		require.Equal(t, map[string]any{
			"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network",
			"attempts": "2", "ready": "true", "empty": "", "cidrs": "[10.0.0.0/24]",
			"null": "<nil>", "nested": "map[message:detail]",
		}, d.Get("state_info"))
	})
}

func TestReadViewClearsMissingComputedValues(t *testing.T) {
	for _, info := range []map[string]any{nil, {}, {"to_project_id": "project"}, {"to_project_id": "", "to_vpc_network": "network"}} {
		client := avngen.NewMockClient(t)
		d := newReadData(t, map[string]any{
			"id": testID, "self_link": "old-link", "state_info": map[string]any{"old": "value"},
		})
		connection := testConnection()
		connection.StateInfo = info
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Empty(t, d.Get("self_link"))
		require.NotContains(t, d.Get("state_info"), "old")
	}
}

func TestReadViewResourceAndDataSourceWarnings(t *testing.T) {
	t.Run("resource", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newReadData(t, map[string]any{"id": testID})
		connection := testConnection()
		connection.State = vpc.VpcPeeringConnectionStateTypePendingPeer
		var diagnostics diag.Diagnostics
		ctx, drain := adapter.WithWarnings(t.Context(), &diagnostics)
		client.EXPECT().VpcGet(ctx, "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(ctx, client, d))
		drain()
		require.Equal(t, testID, d.ID())
		require.Equal(t, 1, diagnostics.WarningsCount())
	})

	t.Run("data source", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(map[string]any{
			"vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
		}))
		require.NoError(t, err)
		connection := testConnection()
		connection.State = vpc.VpcPeeringConnectionStateTypePendingPeer
		var diagnostics diag.Diagnostics
		ctx, drain := adapter.WithWarnings(t.Context(), &diagnostics)
		client.EXPECT().VpcGet(ctx, "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		require.NoError(t, DataSourceOptions.Read(ctx, client, d))
		drain()
		require.Equal(t, testID, d.ID())
		require.Empty(t, diagnostics)
	})
}

func TestReadViewDoesNotDeleteTerminalConnections(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newReadData(t, map[string]any{"id": testID})
	connection := testConnection()
	connection.State = vpc.VpcPeeringConnectionStateTypeInvalidSpecification
	client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
		Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
	require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
	require.Equal(t, "INVALID_SPECIFICATION", d.Get("state"))
}

func TestReadViewMissing(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newReadData(t, map[string]any{"id": testID})
	client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").Return(&vpc.VpcGetOut{}, nil).Once()
	require.ErrorIs(t, ResourceOptions.Read(t.Context(), client, d), adapter.ErrNotFound)
}

func TestDeleteViewUsesLegacyID(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newReadData(t, map[string]any{"id": testID, "gcp_project_id": "stale-project", "peer_vpc": "stale-network"})
	client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "aiven-project", "aiven-vpc", "gcp-project", "peer-network").
		Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func testConnection() vpc.PeeringConnectionOut {
	return vpc.PeeringConnectionOut{PeerCloudAccount: "gcp-project", PeerVpc: "peer-network", State: vpc.VpcPeeringConnectionStateTypeActive}
}

func newReadData(t *testing.T, state map[string]any) adapter.ResourceData {
	t.Helper()
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(state))
	require.NoError(t, err)
	return d
}
