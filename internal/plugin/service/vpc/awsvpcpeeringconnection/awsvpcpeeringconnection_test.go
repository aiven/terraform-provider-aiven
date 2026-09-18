package awsvpcpeeringconnection

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const testID = "project/vpc/123456789012/vpc-peer/eu-west-1"

func TestCreateViewResolvesRegionFromVpcGet(t *testing.T) {
	for _, region := range []string{"", "eu-west-1"} {
		t.Run("region="+region, func(t *testing.T) {
			type request struct {
				method, path, body string
				err                error
			}
			requests := make(chan request, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				requests <- request{r.Method, r.URL.Path, string(body), err}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					// The project POST does not return peer_region or peering_connection_id.
					_, _ = io.WriteString(w, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","state":"APPROVED","state_info":null,"vpc_peering_connection_type":"aws-vpc-peering-connection"}`)
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
			create := <-requests
			require.NoError(t, create.err)
			require.Equal(t, http.MethodPost, create.method)
			require.Equal(t, "/v1/project/project/vpcs/vpc/peering-connections", create.path)
			if region != "" {
				require.JSONEq(t, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer","peer_region":"eu-west-1"}`, create.body)
				require.Equal(t, testID, d.ID())
			} else {
				require.JSONEq(t, `{"peer_cloud_account":"123456789012","peer_vpc":"vpc-peer"}`, create.body)
				require.Equal(t, "project/vpc/123456789012/vpc-peer", d.ID(), "checkpoint before refresh so a failed read cannot orphan the peering")
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
	const regionlessID = "project/vpc/123456789012/vpc-peer"
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
		"id": regionlessID, "aws_vpc_region": "",
	}))
	require.NoError(t, err)
	other := testConnection()
	other.PeerRegion = new("us-east-1")
	client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{other, testConnection()}}, nil).Once()
	require.ErrorIs(t, ResourceOptions.Read(t.Context(), client, d), adapter.ErrMultiple)
	require.Equal(t, regionlessID, d.ID())
}

func TestDeleteViewRegionlessCheckpoint(t *testing.T) {
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(map[string]any{
		"id": "project/vpc/123456789012/vpc-peer", "aws_vpc_region": "",
	}))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer").
		Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func TestReadViewRegionAndCloudID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state map[string]any
		want  string
	}{
		{name: "import", state: map[string]any{"id": testID}, want: "eu-west-1"},
		{name: "SDK state", state: map[string]any{"id": testID, "aws_vpc_region": "eu-west-1"}, want: "eu-west-1"},
		{name: "explicit empty region", state: map[string]any{"id": testID, "aws_vpc_region": ""}},
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
		"id": testID, "aws_vpc_region": "", "aws_account_id": "stale-account", "aws_vpc_id": "stale-vpc",
	}))
	require.NoError(t, err)
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "vpc", "123456789012", "vpc-peer", "eu-west-1").
		Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	require.NoError(t, deleteView(t.Context(), client, d))
}

func TestModifyPlanRegion(t *testing.T) {
	t.Run("replace", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			old     string
			planned any
		}{
			{name: "different region", old: "eu-west-1", planned: "us-east-1"},
			{name: "empty to different region", planned: "us-east-1"},
			{name: "unknown region", old: "eu-west-1", planned: tftypes.UnknownValue},
		} {
			t.Run(tc.name, func(t *testing.T) {
				state := regionPlanValue(t, tc.old)
				plan := regionPlanValue(t, tc.planned)
				req := resource.ModifyPlanRequest{
					Plan: tfsdk.Plan{Raw: plan}, State: tfsdk.State{Raw: state}, Config: tfsdk.Config{Raw: plan},
				}
				rsp := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan}}
				adapter.NewResource(ResourceOptions).(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), req, &rsp)
				require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
				require.Equal(t, path.Paths{path.Root("aws_vpc_region")}, rsp.RequiresReplace)
				require.True(t, plan.Equal(rsp.Plan.Raw), "the configured region must not be rewritten")
			})
		}
	})

	t.Run("no replace", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			old     string
			planned any
		}{
			{name: "unchanged", old: "eu-west-1", planned: "eu-west-1"},
			{name: "clear region without cloud update", old: "eu-west-1", planned: ""},
			{name: "empty stays empty", planned: ""},
			{name: "restore actual region", planned: "eu-west-1"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				state := regionPlanValue(t, tc.old)
				plan := regionPlanValue(t, tc.planned)
				req := resource.ModifyPlanRequest{
					Plan: tfsdk.Plan{Raw: plan}, State: tfsdk.State{Raw: state}, Config: tfsdk.Config{Raw: plan},
				}
				rsp := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan}}
				adapter.NewResource(ResourceOptions).(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), req, &rsp)
				require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
				require.Empty(t, rsp.RequiresReplace)
				require.True(t, plan.Equal(rsp.Plan.Raw), "the configured region must not be rewritten")
			})
		}
	})
}

func regionPlanValue(t *testing.T, region any) tftypes.Value {
	t.Helper()
	typ := resourceSchema(t.Context()).Type().TerraformType(t.Context()).(tftypes.Object)
	values := map[string]any{
		"id": testID, "vpc_id": "project/vpc", "aws_account_id": "123456789012", "aws_vpc_id": "vpc-peer", "aws_vpc_region": region,
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
