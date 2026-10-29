package gcpvpcpeeringconnection

import (
	"context"
	"errors"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
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
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{CloudName: "google-europe-west10"}, nil).Once()
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "aiven-project", "aiven-vpc", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "gcp-project", PeerVpc: "peer-network",
		}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
		require.NoError(t, createView(t.Context(), client, d))
		require.Equal(t, testID, d.ID(), "successful Create must checkpoint the ID before Read")
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
		require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
	})

	t.Run("fails on create error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
			"vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
		}))
		require.NoError(t, err)
		createError := errors.New("create failed")
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{CloudName: "google-europe-west10"}, nil).Once()
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "aiven-project", "aiven-vpc", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "gcp-project", PeerVpc: "peer-network",
		}).Return(&vpc.VpcPeeringConnectionCreateOut{}, createError).Once()
		require.ErrorIs(t, createView(t.Context(), client, d), createError)
		require.Empty(t, d.ID(), "failed Create must not adopt a peering")
	})

	t.Run("does not adopt an existing peering", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
			"vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
		}))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
		require.ErrorContains(t, createView(t.Context(), client, d), "already exists")
		require.Empty(t, d.ID(), "a conflict must not record the existing peering as newly created")
	})
}

func TestReadViewLegacyStateAndImport(t *testing.T) {
	checkRead := func(t *testing.T, d adapter.ResourceData) {
		t.Helper()
		client := avngen.NewMockClient(t)
		connection := testConnection()
		connection.StateInfo = map[string]any{
			"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network",
			"attempts": 2, "ready": true, "empty": "", "cidrs": []any{"10.0.0.0/24"},
			"null": nil, "nested": map[string]any{"message": "detail"},
		}
		other := testConnection()
		other.PeerVpc = "other-network"
		client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, connection}}, nil).Once()
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
	}

	t.Run("import", func(t *testing.T) {
		// Legacy imports accept a fifth component; GCP connections have no region.
		for _, id := range []string{testID, testID + "/"} {
			t.Run(id, func(t *testing.T) {
				r := adapter.NewResource(ResourceOptions).(resource.ResourceWithImportState)
				sch := resourceSchema(t.Context())
				imported := resource.ImportStateResponse{State: tfsdk.State{
					Schema: sch,
					Raw:    tftypes.NewValue(sch.Type().TerraformType(t.Context()), nil),
				}}
				r.ImportState(t.Context(), resource.ImportStateRequest{ID: id}, &imported)
				require.False(t, imported.Diagnostics.HasError(), "%v", imported.Diagnostics)
				d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithState(imported.State))
				require.NoError(t, err)
				checkRead(t, d)
			})
		}
	})

	t.Run("SDK state", func(t *testing.T) {
		d := newReadData(t, map[string]any{
			"id": testID, "vpc_id": "aiven-project/aiven-vpc", "gcp_project_id": "gcp-project", "peer_vpc": "peer-network",
			"state": "PENDING_PEER", "state_info": map[string]any{"old": "value"}, "self_link": "old-link",
		})
		checkRead(t, d)
	})
}

func TestReadViewPreservesKnownSelfLink(t *testing.T) {
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
		require.Equal(t, "old-link", d.Get("self_link"))
		require.NotContains(t, d.Get("state_info"), "old")
	}
}

func TestReadViewPreservesSelfLinkDuringUpdate(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(),
		adapter.WithTestState(map[string]any{
			"id": testID, "self_link": "old-link", "state": "PENDING_PEER",
			"state_info": map[string]any{"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network"},
		}),
		adapter.WithTestPlan(map[string]any{
			"id": testID, "self_link": tftypes.UnknownValue,
		}),
	)
	require.NoError(t, err)
	connection := testConnection()
	connection.State = vpc.VpcPeeringConnectionStateTypeInvalidSpecification
	connection.StateInfo = map[string]any{"type": "google-connect-failed", "message": "overlapping networks"}
	client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
		Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
	require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
	require.Equal(t, "old-link", d.Get("self_link"))
	require.Equal(t, "INVALID_SPECIFICATION", d.Get("state"))
	require.Equal(t, connection.StateInfo, d.Get("state_info"))
}

func TestReadViewWithoutKnownSelfLink(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newReadData(t, map[string]any{"id": testID})
	connection := testConnection()
	connection.State = vpc.VpcPeeringConnectionStateTypeDeleting
	client.EXPECT().VpcGet(t.Context(), "aiven-project", "aiven-vpc").
		Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
	require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
	require.Nil(t, d.Get("self_link"), "a network link that has never been returned must stay null")
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

func TestModifyPlanProjectVPC(t *testing.T) {
	migrated := map[string]any{"id": testID, "project": "aiven-project", "project_vpc_id": "aiven-vpc", "vpc_id": "aiven-project/aiven-vpc"}
	sdk := map[string]any{"id": testID, "vpc_id": "aiven-project/aiven-vpc"}
	separate := map[string]any{"project": "aiven-project", "project_vpc_id": "aiven-vpc"}
	for _, tc := range []struct {
		name    string
		state   map[string]any
		config  map[string]any
		replace bool
	}{
		{name: "switch to separate fields", state: migrated, config: separate},
		{name: "switch to separate fields from SDK state", state: sdk, config: separate},
		{name: "switch to vpc_id", state: migrated, config: map[string]any{"vpc_id": "aiven-project/aiven-vpc"}},
		{name: "change VPC", state: migrated, config: map[string]any{"project": "aiven-project", "project_vpc_id": "other-vpc"}, replace: true},
		{name: "change vpc_id", state: sdk, config: map[string]any{"vpc_id": "aiven-project/other-vpc"}, replace: true},
		{name: "unknown VPC", state: migrated, config: map[string]any{"project": "aiven-project", "project_vpc_id": tftypes.UnknownValue}, replace: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := planValue(t, tc.config, nil)
			planned := map[string]any{"id": testID}
			for name, value := range tc.config {
				planned[name] = value
			}
			plan := planValue(t, planned, tftypes.UnknownValue)
			req := resource.ModifyPlanRequest{
				Plan: tfsdk.Plan{Raw: plan}, State: tfsdk.State{Raw: planValue(t, tc.state, nil)}, Config: tfsdk.Config{Raw: config},
			}
			rsp := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan}}
			adapter.NewResource(ResourceOptions).(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), req, &rsp)
			require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
			if tc.replace {
				require.Equal(t, path.Paths{path.Root("project"), path.Root("project_vpc_id"), path.Root("vpc_id")}, rsp.RequiresReplace)
				return
			}
			require.Empty(t, rsp.RequiresReplace)
			var attributes map[string]tftypes.Value
			require.NoError(t, rsp.Plan.Raw.As(&attributes))
			if tc.state["project"] == nil {
				require.False(t, attributes["create_time"].IsKnown(), "SDK state has no creation timestamp")
			} else {
				require.True(t, attributes["create_time"].IsNull(), "a migrated null timestamp must stay null")
			}
			for name, value := range migrated {
				var planned string
				require.NoError(t, attributes[name].As(&planned), name)
				require.Equal(t, value, planned, name)
			}
		})
	}
}

func TestMigrationWithoutRefreshComputedFields(t *testing.T) {
	// An import can record APPROVED before the API exposes the network link.
	// Without refresh, Framework receives an unchanged proposed plan and does not
	// mark computed attributes unknown before our selector migration runs.
	previousTime := time.Date(2026, time.October, 23, 10, 0, 0, 0, time.UTC)
	sch := resourceSchema(t.Context())
	state := tfsdk.State{Raw: planValue(t, map[string]any{
		"id": testID, "vpc_id": "aiven-project/aiven-vpc", "state": "APPROVED",
		"create_time": previousTime.Format(time.RFC3339),
	}, nil), Schema: sch}
	config := tfsdk.Config{Raw: planValue(t, map[string]any{"vpc_id": "aiven-project/aiven-vpc"}, nil)}
	plan := tfsdk.Plan{Raw: state.Raw, Schema: sch}

	client := avngen.NewMockClient(t)
	options := ResourceOptions
	options.Read = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
		return readView(ctx, client, d)
	}
	r := adapter.NewResource(options)
	planned := resource.ModifyPlanResponse{Plan: plan}
	r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), resource.ModifyPlanRequest{
		State: state, Plan: plan, Config: config,
	}, &planned)
	require.False(t, planned.Diagnostics.HasError(), "%v", planned.Diagnostics)
	require.Empty(t, planned.RequiresReplace)

	connection := testConnection()
	// An external re-request preserves the ID but changes the creation timestamp.
	connection.CreateTime = previousTime.Add(time.Hour)
	connection.State = vpc.VpcPeeringConnectionStateTypePendingPeer
	connection.StateInfo = map[string]any{"to_project_id": "aiven-gcp-project", "to_vpc_network": "aiven-network"}
	client.EXPECT().VpcGet(mock.Anything, "aiven-project", "aiven-vpc").
		Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
	updated := resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	r.Update(t.Context(), resource.UpdateRequest{State: state, Plan: planned.Plan, Config: config}, &updated)
	require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)

	var link string
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("self_link"), &link).HasError())
	require.Equal(t, "https://www.googleapis.com/compute/v1/projects/aiven-gcp-project/global/networks/aiven-network", link)
	var timestamp string
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("create_time"), &timestamp).HasError())
	require.Equal(t, connection.CreateTime.Format(time.RFC3339), timestamp)
	var plannedAttributes, updatedAttributes map[string]tftypes.Value
	require.NoError(t, planned.Plan.Raw.As(&plannedAttributes))
	require.NoError(t, updated.State.Raw.As(&updatedAttributes))
	for _, name := range []string{"self_link", "create_time"} {
		plannedValue := plannedAttributes[name]
		require.True(t, !plannedValue.IsKnown() || plannedValue.Equal(updatedAttributes[name]),
			"migration must not change a known planned %s: planned %s, got %s", name, plannedValue, updatedAttributes[name])
	}
}

// planValue sets the peer fields and values; unset VPC fields get missing.
func planValue(t *testing.T, values map[string]any, missing any) tftypes.Value {
	t.Helper()
	typ := resourceSchema(t.Context()).Type().TerraformType(t.Context()).(tftypes.Object)
	all := map[string]any{"gcp_project_id": "gcp-project", "peer_vpc": "peer-network"}
	for _, name := range []string{"project", "project_vpc_id", "vpc_id"} {
		all[name] = missing
	}
	for name, value := range values {
		all[name] = value
	}
	attributes := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attributeType := range typ.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, all[name])
	}
	return tftypes.NewValue(typ, attributes)
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
