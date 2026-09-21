package azurevpcpeeringconnection

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const testID = "project/vpc/subscription/network"

func TestCreateViewResourceGroup(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(testConfig()))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "project", "vpc", &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: new("group"),
		PeerAzureAppId: new("app"), PeerAzureTenantId: new("tenant"),
	}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
	require.NoError(t, createView(t.Context(), client, d))
	require.Equal(t, testID, d.ID(), "resource group must not become a new ID component")
}

func TestReadViewLegacyStateAndImport(t *testing.T) {
	t.Run("import", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{"id": testID}))
		require.NoError(t, err)
		connection := testConnection()
		connection.PeeringConnectionId = new("aiven-api-id")
		connection.StateInfo = map[string]any{"message": "ready", "attempts": 2}
		decoy := connection
		decoy.PeerCloudAccount = "other-subscription"
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{decoy, connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
		for name, value := range testConfig() {
			require.Equal(t, value, d.Get(name), name)
		}
		require.Empty(t, d.Get("peering_connection_id"), "the API ID is not the SDK's reserved cloud ID")
		require.Equal(t, "2", d.Get("state_info.attempts"))
	})

	t.Run("SDK state", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
			"id": testID, "peer_resource_group": "group", "peering_connection_id": "",
		}))
		require.NoError(t, err)
		connection := testConnection()
		connection.PeeringConnectionId = new("aiven-api-id")
		connection.StateInfo = map[string]any{"message": "ready", "attempts": 2}
		otherSubscription := connection
		otherSubscription.PeerCloudAccount = "other-subscription"
		otherGroup := connection
		otherGroup.PeerResourceGroup = "other-group"
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{otherSubscription, otherGroup, connection}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
		for name, value := range testConfig() {
			require.Equal(t, value, d.Get(name), name)
		}
		require.Empty(t, d.Get("peering_connection_id"), "the API ID is not the SDK's reserved cloud ID")
		require.IsType(t, "", d.Get("peering_connection_id"), "preserve the legacy empty string")
		require.Equal(t, "2", d.Get("state_info.attempts"))
	})
}

func TestDatasourceReadViewMatchesAllKeys(t *testing.T) {
	fields := []string{"account", "network", "group", "app", "tenant"}
	connections := make([]vpc.PeeringConnectionOut, 0, len(fields))
	for _, field := range fields {
		decoy := testConnection()
		switch field {
		case "account":
			decoy.PeerCloudAccount = "wrong"
		case "network":
			decoy.PeerVpc = "wrong"
		case "group":
			decoy.PeerResourceGroup = "wrong"
		case "app":
			decoy.PeerAzureAppId = "wrong"
		case "tenant":
			decoy.PeerAzureTenantId = "wrong"
		}
		connections = append(connections, decoy)
	}

	t.Run("found", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(testConfig()))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: append(connections, testConnection())}, nil).Once()
		require.NoError(t, DataSourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
	})

	t.Run("not found", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(testConfig()))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: connections}, nil).Once()
		require.ErrorIs(t, DataSourceOptions.Read(t.Context(), client, d), adapter.ErrNotFound)
	})
}

func TestReadViewRejectsAmbiguousImport(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{"id": testID}))
	require.NoError(t, err)
	other := testConnection()
	other.PeerResourceGroup = "other-group"
	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, testConnection()}}, nil).Once()
	require.ErrorIs(t, ResourceOptions.Read(t.Context(), client, d), adapter.ErrMultiple)
	require.Equal(t, testID, d.ID())
}

func TestDeleteViewResourceGroup(t *testing.T) {
	for _, group := range []string{"", "group"} {
		t.Run("group="+group, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
				"id": testID, "peer_resource_group": group, "azure_subscription_id": "stale", "vnet_name": "stale",
			}))
			require.NoError(t, err)
			if group == "" {
				client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
			}
			client.EXPECT().VpcPeeringConnectionWithResourceGroupDelete(t.Context(), "project", "vpc", "subscription", "group", "network").
				Return(&vpc.VpcPeeringConnectionWithResourceGroupDeleteOut{}, nil).Once()
			require.NoError(t, deleteView(t.Context(), client, d))
		})
	}
}

func TestCreateViewOmitsEmptyAzureFields(t *testing.T) {
	client := avngen.NewMockClient(t)
	config := testConfig()
	for _, name := range []string{"peer_resource_group", "peer_azure_app_id", "peer_azure_tenant_id"} {
		config[name] = ""
	}
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(config))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "project", "vpc", &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: "subscription", PeerVpc: "network",
	}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
	require.NoError(t, createView(t.Context(), client, d))
}

func testConfig() map[string]any {
	return map[string]any{
		"vpc_id": "project/vpc", "azure_subscription_id": "subscription", "vnet_name": "network",
		"peer_resource_group": "group", "peer_azure_app_id": "app", "peer_azure_tenant_id": "tenant",
	}
}

func testConnection() vpc.PeeringConnectionOut {
	return vpc.PeeringConnectionOut{
		PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: "group",
		PeerAzureAppId: "app", PeerAzureTenantId: "tenant", State: vpc.VpcPeeringConnectionStateTypeActive,
	}
}
