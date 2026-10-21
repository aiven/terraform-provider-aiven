package awsvpcpeeringconnection

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testID   = "project/vpc/123456789012/vpc-peer"
	legacyID = testID + "/eu-west-1"
)

func TestCreateViewResolvesRegionFromVpcGet(t *testing.T) {
	for _, region := range []string{"", "eu-west-1"} {
		t.Run("region="+region, func(t *testing.T) {
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
			if region != "" {
				require.JSONEq(t, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","peer_region":"eu-west-1"}`, create.body)
			} else {
				require.JSONEq(t, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer"}`, create.body)
			}
			require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
			read := <-requests
			require.NoError(t, read.err)
			require.Equal(t, http.MethodGet, read.method)
			require.Equal(t, "/v1/project/project/vpcs/vpc", read.path)
			require.Equal(t, testID, d.ID())
			require.Equal(t, region, d.Get("aws_vpc_region"))
			require.Equal(t, "pcx-current", d.Get("aws_vpc_peering_connection_id"))
			require.Equal(t, "<nil>", d.Get("state_info.future_detail"))
		})
	}
}

func TestReadViewRejectsAmbiguousRegion(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
		"id": testID, "aws_vpc_region": "",
	}))
	require.NoError(t, err)
	other := testConnection()
	other.PeerRegion = new("us-east-1")
	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, testConnection()}}, nil).Once()
	require.ErrorIs(t, ResourceOptions.Read(t.Context(), client, d), adapter.ErrMultiple)
	require.Equal(t, testID, d.ID())
}

func TestDeleteViewResolvesEmptyRegion(t *testing.T) {
	newData := func(t *testing.T) adapter.ResourceData {
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
			"id": testID, "aws_vpc_region": "",
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

	t.Run("regionless connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		connection := testConnection()
		connection.PeerRegion = nil
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer").
			Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
		require.NoError(t, deleteView(t.Context(), client, newData(t)))
	})

	t.Run("already deleted", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{}, nil).Once()
		require.ErrorIs(t, deleteView(t.Context(), client, newData(t)), adapter.ErrNotFound)
	})
}

func TestReadViewImport(t *testing.T) {
	for _, tc := range []struct {
		name, network string
	}{
		{name: "composed ID", network: "vpc-peer"},
		{name: "legacy ID with region", network: "vpc-peer/eu-west-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
				"project": "project", "project_vpc_id": "vpc", "aws_account_id": "123456789012", "aws_vpc_id": tc.network,
			}))
			require.NoError(t, err)
			connections := []vpc.PeeringConnectionOut{testConnection()}
			if tc.network != "vpc-peer" {
				other := testConnection()
				other.PeerRegion = new("us-east-1")
				connections = append(connections, other)
			}
			client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: connections}, nil).Once()
			require.NoError(t, ResourceOptions.Read(t.Context(), client, d))
			require.Equal(t, testID, d.ID())
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
		{name: "legacy ID", state: map[string]any{"id": legacyID}, want: "eu-west-1"},
		{name: "SDK state", state: map[string]any{"id": legacyID, "aws_vpc_region": "eu-west-1"}, want: "eu-west-1"},
		{name: "explicit empty region", state: map[string]any{"id": legacyID, "aws_vpc_region": ""}},
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
			require.Equal(t, testID, d.ID())
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

func TestDatasourceReadViewExactRegion(t *testing.T) {
	t.Run("empty region", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		d, err := adapter.NewResourceData(datasourceSchemaInternal(), idFields(), adapter.WithIsDataSource(), adapter.WithTestConfig(map[string]any{
			"vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": "",
		}))
		require.NoError(t, err)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
		require.ErrorIs(t, DataSourceOptions.Read(t.Context(), client, d), adapter.ErrNotFound)
	})

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
		"id": legacyID, "aws_vpc_region": "", "aws_account_id": "stale-account", "aws_vpc_id": "stale-vpc",
	}))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer", "eu-west-1").
		Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func TestModifyPlanRegion(t *testing.T) {
	for _, tc := range []struct {
		name    string
		id      string
		old     string
		planned any
		replace bool
	}{
		{name: "different region", id: legacyID, old: "eu-west-1", planned: "us-east-1", replace: true},
		{name: "empty to different region", id: legacyID, planned: "us-east-1", replace: true},
		{name: "unknown region", id: legacyID, old: "eu-west-1", planned: tftypes.UnknownValue, replace: true},
		{name: "unchanged", id: legacyID, old: "eu-west-1", planned: "eu-west-1"},
		{name: "clear region without cloud update", id: legacyID, old: "eu-west-1", planned: ""},
		{name: "empty stays empty", id: legacyID, planned: ""},
		{name: "restore actual region", id: legacyID, planned: "eu-west-1"},
		{name: "different region in state", id: testID, old: "eu-west-1", planned: "us-east-1", replace: true},
		{name: "empty to different looked-up region", id: testID, planned: "us-east-1", replace: true},
		{name: "restore looked-up region", id: testID, planned: "eu-west-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			if tc.id == testID && tc.old == "" {
				client.EXPECT().VpcGet(mock.Anything, "project", "vpc").
					Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{testConnection()}}, nil).Once()
			}
			r := adapter.NewResource(ResourceOptions)
			r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: testProviderData{client}}, &resource.ConfigureResponse{})

			state := regionPlanValue(t, tc.id, tc.old)
			plan := regionPlanValue(t, tc.id, tc.planned)
			req := resource.ModifyPlanRequest{
				Plan: tfsdk.Plan{Raw: plan}, State: tfsdk.State{Raw: state}, Config: tfsdk.Config{Raw: plan},
			}
			rsp := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan}}
			r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), req, &rsp)
			require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
			if tc.replace {
				require.Equal(t, path.Paths{path.Root("aws_vpc_region")}, rsp.RequiresReplace)
			} else {
				require.Empty(t, rsp.RequiresReplace)
			}
			require.True(t, plan.Equal(rsp.Plan.Raw), "the configured region must not be rewritten")
		})
	}
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
