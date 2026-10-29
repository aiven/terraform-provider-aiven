package awsvpcpeeringconnection

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	fourPartID = "project/vpc/123456789012/vpc-peer"
	testID     = fourPartID + "/eu-west-1"
)

func TestCreateViewWithIncompleteResponse(t *testing.T) {
	const region = "eu-west-1"
	type request struct {
		method, path, body string
		err                error
	}
	requests := make(chan request, 3)
	var created atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		requests <- request{r.Method, r.URL.Path, string(body), err}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			created.Store(true)
			// The project POST does not return peer_region or peering_connection_id.
			_, _ = io.WriteString(w, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","state":"APPROVED","state_info":null,"vpc_peering_connection_type":"aws-vpc-peering-connection"}`)
			return
		}
		if !created.Load() {
			_, _ = io.WriteString(w, `{"cloud_name":"aws-eu-west-1","project_vpc_id":"vpc","peering_connections":[]}`)
			return
		}
		_, _ = io.WriteString(w, `{"cloud_name":"aws-eu-west-1","project_vpc_id":"vpc","peering_connections":[{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","peer_region":"eu-west-1","peer_resource_group":null,"peer_azure_app_id":null,"peer_azure_tenant_id":null,"state":"PENDING_PEER","state_info":{"type":"aws-accept-peering-connection-request","message":"Accept the peering request","aws_vpc_peering_connection_id":"pcx-current","future_detail":null},"vpc_peering_connection_type":"aws-vpc-peering-connection"}]}`)
	}))
	defer server.Close()
	client, err := avngen.NewClient(avngen.TokenOpt("test"), avngen.HostOpt(server.URL), avngen.DoerOpt(server.Client()), avngen.DebugOpt(false))
	require.NoError(t, err)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
		"vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": region,
	}))
	require.NoError(t, err)
	require.NoError(t, createView(t.Context(), client, d))
	lookup := <-requests
	require.NoError(t, lookup.err)
	require.Equal(t, http.MethodGet, lookup.method)
	require.Equal(t, "/v1/project/project/vpcs/vpc", lookup.path)
	create := <-requests
	require.NoError(t, create.err)
	require.Equal(t, http.MethodPost, create.method)
	require.Equal(t, "/v1/project/project/vpcs/vpc/peering-connections", create.path)
	require.JSONEq(t, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","peer_region":"eu-west-1"}`, create.body)
	require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
	read := <-requests
	require.NoError(t, read.err)
	require.Equal(t, http.MethodGet, read.method)
	require.Equal(t, "/v1/project/project/vpcs/vpc", read.path)
	require.Equal(t, testID, d.ID())
	require.Equal(t, region, d.Get("aws_vpc_region"))
	require.Equal(t, "pcx-current", d.Get("aws_vpc_peering_connection_id"))
	require.Equal(t, "<nil>", d.Get("state_info.future_detail"))
}

func TestReadViewRejectsAmbiguousRegion(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
		"id": fourPartID, "aws_vpc_region": "",
	}))
	require.NoError(t, err)
	other := testConnection()
	other.PeerRegion = new("us-east-1")
	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, testConnection()}}, nil).Once()
	require.ErrorIs(t, ResourceOptions.Read(t.Context(), client, d), adapter.ErrMultiple)
	require.Equal(t, fourPartID, d.ID())
}

func TestCreateViewWithOtherRegion(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(map[string]any{
		"vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": "eu-west-1",
	}))
	require.NoError(t, err)
	other := testConnection()
	other.PeerRegion = new("us-east-1")
	other.State = vpc.VpcPeeringConnectionStateTypeInvalidSpecification
	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{
		CloudName: "aws-eu-west-1", PeeringConnections: []vpc.PeeringConnectionOut{other},
	}, nil).Once()
	client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "project", "vpc", &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: "123456789012", PeerVpc: "vpc-peer", PeerRegion: new("eu-west-1"),
	}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
	require.NoError(t, createView(t.Context(), client, d))
	require.Equal(t, testID, d.ID())

	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{other, testConnection()},
	}, nil).Once()
	require.NoError(t, readView(t.Context(), client, d))
	require.Equal(t, testID, d.ID())
	require.Equal(t, "eu-west-1", d.Get("aws_vpc_region"))

	client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer", "eu-west-1").
		Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func TestDeleteViewResolvesEmptyRegion(t *testing.T) {
	newData := func(t *testing.T) adapter.ResourceData {
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
			"id": fourPartID, "aws_vpc_region": "",
		}))
		require.NoError(t, err)
		return d
	}

	t.Run("regional connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		connection := testConnection()
		connection.PeerRegion = new("us-east-1")
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer", "us-east-1").
			Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
		require.NoError(t, deleteView(t.Context(), client, newData(t)))
	})

	t.Run("regionless connection after read", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d := newData(t)
		connection := testConnection()
		// Defensive coverage: ordinary AWS connections always have an API region.
		connection.PeerRegion = nil
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Twice()
		client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer").
			Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
		require.NoError(t, readView(t.Context(), client, d))
		require.NoError(t, deleteView(t.Context(), client, d))
	})

	t.Run("already deleted", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{}, nil).Once()
		require.ErrorIs(t, deleteView(t.Context(), client, newData(t)), adapter.ErrNotFound)
	})
}

func TestReadViewImport(t *testing.T) {
	for _, id := range []string{fourPartID, testID} {
		t.Run(id, func(t *testing.T) {
			client := avngen.NewMockClient(t)
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
			connections := []vpc.PeeringConnectionOut{testConnection()}
			if id == testID {
				other := testConnection()
				other.PeerRegion = new("us-east-1")
				connections = append(connections, other)
			}
			client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: connections}, nil).Once()
			require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
			require.Equal(t, id, d.ID())
			require.Equal(t, "vpc-peer", d.Get("aws_vpc_id"))
			require.Equal(t, "eu-west-1", d.Get("aws_vpc_region"))
		})
	}
}

func TestReadViewRegionAndCloudID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state map[string]any
		want  string
	}{
		{name: "legacy ID", state: map[string]any{"id": testID}, want: "eu-west-1"},
		{name: "SDK state", state: map[string]any{"id": testID, "aws_vpc_region": "eu-west-1"}, want: "eu-west-1"},
		{name: "four-part SDK import", state: map[string]any{"id": fourPartID, "aws_vpc_region": "eu-west-1"}, want: "eu-west-1"},
		{name: "empty region in old state", state: map[string]any{"id": testID, "aws_vpc_region": ""}, want: "eu-west-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(tc.state))
			require.NoError(t, err)
			connection := testConnection()
			connection.StateInfo = map[string]any{"aws_vpc_peering_connection_id": "pcx-current", "attempts": 2}
			client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{
				{PeerCloudAccount: "123456789012", PeerVpc: "vpc-peer", PeerRegion: new("us-east-1")}, connection,
			}}, nil).Once()
			require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
			require.Equal(t, tc.state["id"], d.ID())
			require.Equal(t, "project/vpc", d.Get("vpc_id"))
			require.Equal(t, tc.want, d.Get("aws_vpc_region"))
			require.Equal(t, "pcx-current", d.Get("aws_vpc_peering_connection_id"))
			require.Equal(t, "2", d.Get("state_info.attempts"))
			connection.StateInfo = nil
			client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
			require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
			require.Empty(t, d.Get("aws_vpc_peering_connection_id"))
			require.Empty(t, d.Get("state_info"))
		})
	}
}

func TestReadAfterRegionalConnectionAdded(t *testing.T) {
	connection := testConnection()
	r, created, client := createConnection(t, connection)
	other := connection
	other.PeerRegion = new("us-east-1")
	other.State = vpc.VpcPeeringConnectionStateTypeInvalidSpecification
	client.EXPECT().VpcGet(mock.Anything, "project", "vpc").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{connection, other},
	}, nil).Once()
	read := resource.ReadResponse{State: created.State}
	r.Read(t.Context(), resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "another region must not break this resource's refresh: %v", read.Diagnostics)
	var state string
	require.False(t, read.State.GetAttribute(t.Context(), path.Root("state"), &state).HasError())
	require.Equal(t, "ACTIVE", state)
}

func TestDatasourceReadViewExactRegion(t *testing.T) {
	t.Run("matching region", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(map[string]any{
			"vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": "eu-west-1",
		}))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
		require.NoError(t, DataSourceOptions.Read(t.Context(), client, d))
		require.Equal(t, testID, d.ID())
	})

	t.Run("different region", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(map[string]any{
			"vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": "us-east-1",
		}))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
		require.ErrorIs(t, DataSourceOptions.Read(t.Context(), client, d), adapter.ErrNotFound)
	})
}

func TestDeleteViewUsesRegionFromID(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
		"id": testID, "aws_vpc_region": "", "aws_account_id": "stale-account", "aws_vpc_id": "stale-vpc",
	}))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer", "eu-west-1").
		Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func TestDeleteAfterReplacement(t *testing.T) {
	check := func(t *testing.T, originalVisible bool) {
		t.Helper()
		original := testConnection()
		original.State = vpc.VpcPeeringConnectionStateTypeInvalidSpecification
		r, created, client := createConnection(t, original)
		replacement := testConnection()
		replacement.PeerRegion = new("us-east-1")
		replacement.State = vpc.VpcPeeringConnectionStateTypePendingPeer
		connections := []vpc.PeeringConnectionOut{replacement}
		if originalVisible {
			connections = append(connections, original)
		}
		client.EXPECT().VpcGet(mock.Anything, "project", "vpc").RunAndReturn(
			func(context.Context, string, string) (*vpc.VpcGetOut, error) {
				return &vpc.VpcGetOut{PeeringConnections: connections}, nil
			},
		)
		var deletedRegions []string
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(mock.Anything, "project", "vpc", "123456789012", "vpc-peer", mock.Anything).
			RunAndReturn(func(_ context.Context, _, _, _, _, region string) (*vpc.VpcPeeringConnectionWithRegionDeleteOut, error) {
				deletedRegions = append(deletedRegions, region)
				for i, connection := range connections {
					if *connection.PeerRegion == region {
						connections = append(connections[:i], connections[i+1:]...)
						return &vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil
					}
				}
				return nil, avngen.Error{Status: http.StatusNotFound}
			}).Maybe()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		deleted := resource.DeleteResponse{State: created.State}
		r.Delete(ctx, resource.DeleteRequest{State: created.State}, &deleted)
		require.NotContains(t, deletedRegions, "us-east-1", "deleting the deposed instance must not delete its replacement")
		require.False(t, deleted.Diagnostics.HasError(), "deleting the deposed instance must use its original region: %v", deleted.Diagnostics)
		require.Equal(t, []vpc.PeeringConnectionOut{replacement}, connections)
	}
	t.Run("original still visible", func(t *testing.T) { check(t, true) })
	t.Run("original no longer visible", func(t *testing.T) { check(t, false) })
}

func TestMigrationWithoutRefresh(t *testing.T) {
	check := func(t *testing.T, createTime time.Time, previousState string, currentState vpc.VpcPeeringConnectionStateType) {
		t.Helper()
		client := avngen.NewMockClient(t)
		connection := testConnection()
		connection.CreateTime = createTime
		connection.State = currentState
		connection.StateInfo = map[string]any{"aws_vpc_peering_connection_id": "pcx-test"}
		if currentState != vpc.VpcPeeringConnectionStateTypeActive {
			connection.StateInfo = nil
		}
		other := testConnection()
		other.PeerRegion = new("us-east-1")
		r := adapter.NewResource(ResourceOptions)
		r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: testProviderData{client}}, &resource.ConfigureResponse{})
		sch := resourceSchema(t.Context())
		state := tfsdk.State{Raw: regionPlanValue(t, testID, "eu-west-1"), Schema: sch}
		// SDK state has neither the separate VPC fields nor create_time.
		require.False(t, state.SetAttribute(t.Context(), path.Root("project"), types.StringNull()).HasError())
		require.False(t, state.SetAttribute(t.Context(), path.Root("project_vpc_id"), types.StringNull()).HasError())
		require.False(t, state.SetAttribute(t.Context(), path.Root("state"), previousState).HasError())
		require.False(t, state.SetAttribute(t.Context(), path.Root("aws_vpc_peering_connection_id"), "pcx-test").HasError())
		previousInfo := map[string]string{"aws_vpc_peering_connection_id": "pcx-test"}
		if previousState == "PENDING_PEER" {
			previousInfo["type"] = "aws-accept-peering-connection-request"
			previousInfo["message"] = "Accept the peering request"
		}
		require.False(t, state.SetAttribute(t.Context(), path.Root("state_info"), previousInfo).HasError())
		config := tfsdk.Config{Raw: regionPlanValue(t, "", "eu-west-1")}
		var configAttributes map[string]tftypes.Value
		require.NoError(t, config.Raw.As(&configAttributes))
		for name, field := range resourceSchemaInternal().Properties {
			if field.Computed && name != "vpc_id" {
				configAttributes[name] = tftypes.NewValue(configAttributes[name].Type(), nil)
			}
		}
		config.Raw = tftypes.NewValue(config.Raw.Type(), configAttributes)

		// With refresh disabled and unchanged configuration, the proposed plan equals SDK state.
		plan := tfsdk.Plan{Raw: state.Raw, Schema: sch}
		planned := resource.ModifyPlanResponse{Plan: plan}
		r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), resource.ModifyPlanRequest{
			Plan: plan, State: state, Config: config,
		}, &planned)
		require.False(t, planned.Diagnostics.HasError(), "%v", planned.Diagnostics)
		require.Empty(t, planned.RequiresReplace)
		var plannedAttributes map[string]tftypes.Value
		require.NoError(t, planned.Plan.Raw.As(&plannedAttributes))
		require.True(t, plannedAttributes["id"].IsKnown(), "the SDK ID must remain unchanged")
		require.False(t, plannedAttributes["create_time"].IsKnown(), "the new timestamp must be read during apply")
		require.False(t, plannedAttributes["state"].IsKnown(), "the peering status may have changed since the last refresh")
		require.False(t, plannedAttributes["state_info"].IsKnown(), "the peering details may have changed since the last refresh")

		client.EXPECT().VpcGet(mock.Anything, "project", "vpc").
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, connection}}, nil).Once()
		updated := resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
		// A state migration must not wait for a rejected or transient connection.
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		r.Update(ctx, resource.UpdateRequest{Plan: planned.Plan, State: state, Config: config}, &updated)
		require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
		var attributes map[string]tftypes.Value
		require.NoError(t, updated.State.Raw.As(&attributes))
		for name, value := range plannedAttributes {
			if value.IsKnown() {
				require.True(t, value.Equal(attributes[name]), "%s changed from its known planned value", name)
			}
		}
		var id string
		require.NoError(t, attributes["id"].As(&id))
		require.Equal(t, testID, id)
		var actualState string
		require.NoError(t, attributes["state"].As(&actualState))
		require.Equal(t, string(currentState), actualState)
		var currentInfo map[string]string
		require.False(t, updated.State.GetAttribute(t.Context(), path.Root("state_info"), &currentInfo).HasError())
		if currentState == vpc.VpcPeeringConnectionStateTypeActive {
			require.Equal(t, map[string]string{"aws_vpc_peering_connection_id": "pcx-test"}, currentInfo)
		} else {
			require.Empty(t, currentInfo)
			require.True(t, attributes["aws_vpc_peering_connection_id"].IsNull())
		}
		if createTime.IsZero() {
			require.True(t, attributes["create_time"].IsNull())
		} else {
			var timestamp string
			require.NoError(t, attributes["create_time"].As(&timestamp))
			require.Equal(t, createTime.Format(time.RFC3339), timestamp)
		}

		// Migration must not cause another update, even when the API timestamp is null.
		plan = tfsdk.Plan{Raw: updated.State.Raw, Schema: sch}
		next := resource.ModifyPlanResponse{Plan: plan}
		r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), resource.ModifyPlanRequest{
			Plan: plan, State: updated.State, Config: config,
		}, &next)
		require.False(t, next.Diagnostics.HasError(), "%v", next.Diagnostics)
		require.Empty(t, next.RequiresReplace)
		require.True(t, updated.State.Raw.Equal(next.Plan.Raw), "the next plan must be empty")
	}

	t.Run("timestamp returned by API", func(t *testing.T) {
		check(t, time.Date(2026, time.October, 21, 13, 0, 12, 0, time.UTC), "ACTIVE", vpc.VpcPeeringConnectionStateTypeActive)
	})
	t.Run("null timestamp returned by API", func(t *testing.T) {
		check(t, time.Time{}, "ACTIVE", vpc.VpcPeeringConnectionStateTypeActive)
	})
	t.Run("peering became active before migration", func(t *testing.T) {
		check(t, time.Date(2026, time.October, 21, 13, 0, 12, 0, time.UTC), "PENDING_PEER", vpc.VpcPeeringConnectionStateTypeActive)
	})
	t.Run("peering was rejected before migration", func(t *testing.T) {
		check(t, time.Time{}, "PENDING_PEER", vpc.VpcPeeringConnectionStateTypeRejectedByPeer)
	})
	t.Run("peering is transient during migration", func(t *testing.T) {
		check(t, time.Time{}, "PENDING_PEER", vpc.VpcPeeringConnectionStateTypeApproved)
	})
}

type testProviderData struct{ client avngen.Client }

func (testProviderData) GetClient() *aiven.Client      { return nil }
func (p testProviderData) GetGenClient() avngen.Client { return p.client }

func regionPlanValue(t *testing.T, id string, region any) tftypes.Value {
	t.Helper()
	typ := resourceSchema(t.Context()).Type().TerraformType(t.Context()).(tftypes.Object)
	values := map[string]any{
		"id": id, "project": "project", "project_vpc_id": "vpc", "vpc_id": "project/vpc",
		"aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": region,
	}
	attributes := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attributeType := range typ.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, values[name])
	}
	return tftypes.NewValue(typ, attributes)
}

func testConnection() vpc.PeeringConnectionOut {
	return vpc.PeeringConnectionOut{
		PeerCloudAccount: "123456789012", PeerVpc: "vpc-peer", PeerRegion: new("eu-west-1"), State: vpc.VpcPeeringConnectionStateTypeActive,
	}
}

func createConnection(t *testing.T, connection vpc.PeeringConnectionOut) (resource.Resource, resource.CreateResponse, *avngen.MockClient) {
	t.Helper()
	client := avngen.NewMockClient(t)
	r := adapter.NewResource(ResourceOptions)
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: testProviderData{client}}, &resource.ConfigureResponse{})
	sch := resourceSchema(t.Context())
	config := tfsdk.Config{Raw: regionPlanValue(t, "", "eu-west-1")}
	plan := tfsdk.Plan{Raw: config.Raw, Schema: sch}
	require.False(t, plan.SetAttribute(t.Context(), path.Root("id"), types.StringUnknown()).HasError())
	client.EXPECT().VpcGet(mock.Anything, "project", "vpc").Return(&vpc.VpcGetOut{CloudName: "aws-eu-west-1"}, nil).Once()
	client.EXPECT().VpcPeeringConnectionCreate(mock.Anything, "project", "vpc", &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: "123456789012", PeerVpc: "vpc-peer", PeerRegion: new("eu-west-1"),
	}).Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once()
	client.EXPECT().VpcGet(mock.Anything, "project", "vpc").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{connection},
	}, nil).Once()
	created := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: sch}}
	r.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &created)
	if connection.State == vpc.VpcPeeringConnectionStateTypeInvalidSpecification {
		// A wrong region fails readiness, but its checkpointed state must
		// still identify this instance when Terraform later replaces it.
		require.True(t, created.Diagnostics.HasError())
		require.Contains(t, created.Diagnostics.Errors()[0].Detail(), "specification is invalid")
	} else {
		require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
	}
	var id string
	require.False(t, created.State.GetAttribute(t.Context(), path.Root("id"), &id).HasError())
	require.Equal(t, testID, id)
	return r, created, client
}
