package gcpprivatelinkconnectionapproval

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testProject = "test-project"
	testService = "test-service"
	testPSC     = "50079288774361602"
	testIP      = "10.0.0.2"
)

func connectionWithoutIP(state privatelink.ConnectionStateType) privatelink.ServicePrivatelinkGoogleConnectionListOut {
	return privatelink.ServicePrivatelinkGoogleConnectionListOut{
		PrivatelinkConnectionId: "plc1", PscConnectionId: testPSC, State: state,
	}
}

func connectionWithIP(state privatelink.ConnectionStateType, ip string) privatelink.ServicePrivatelinkGoogleConnectionListOut {
	result := connectionWithoutIP(state)
	result.UserIPAddress = ip
	return result
}

func approvedResponse() *privatelink.ServicePrivatelinkGoogleConnectionCreateOut {
	return &privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
		PrivatelinkConnectionId: "plc1", PscConnectionId: testPSC,
		State: privatelink.ServicePrivatelinkGoogleConnectionStateTypeUserApproved, UserIPAddress: testIP,
	}
}

func newPlan(t *testing.T, selector string) adapter.ResourceData {
	t.Helper()
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestPlan(map[string]any{
		"project": testProject, "service_name": testService, "user_ip_address": testIP, "psc_connection_id": selector,
	}))
	require.NoError(t, err)
	return d
}

func newPlanWithoutSelector(t *testing.T) adapter.ResourceData {
	t.Helper()
	d := newPlan(t, "")
	require.NoError(t, d.Set("psc_connection_id", nil))
	return d
}

func TestCreateApproval(t *testing.T) {
	t.Run("new connection", func(t *testing.T) {
		t.Run("single connection without a selector", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1",
				&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: testIP}).Return(approvedResponse(), nil).Once()
			d := newPlanWithoutSelector(t)
			require.NoError(t, createView(t.Context(), client, d))
			require.Equal(t, testProject+"/"+testService, d.ID())
			require.Equal(t, "plc1", d.Get("privatelink_connection_id"))
			require.Equal(t, testPSC, d.Get("psc_connection_id"))
			require.Equal(t, testIP, d.Get("user_ip_address"))
			require.Equal(t, "user-approved", d.Get("state"))
		})

		t.Run("selected connection among multiple", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{
					connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval),
					{PrivatelinkConnectionId: "plc-other", PscConnectionId: "other", State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.3"},
				}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1",
				&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: testIP}).Return(approvedResponse(), nil).Once()
			d := newPlan(t, testPSC)
			require.NoError(t, createView(t.Context(), client, d))
			require.Equal(t, testProject+"/"+testService, d.ID())
			require.Equal(t, "plc1", d.Get("privatelink_connection_id"))
			require.Equal(t, testPSC, d.Get("psc_connection_id"))
			require.Equal(t, testIP, d.Get("user_ip_address"))
			require.Equal(t, "user-approved", d.Get("state"))
		})
	})

	t.Run("existing connection", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				state privatelink.ConnectionStateType
			}{
				{name: "already user-approved", state: privatelink.ConnectionStateTypeUserApproved},
				{name: "already active", state: privatelink.ConnectionStateTypeActive},
			} {
				t.Run(tc.name, func(t *testing.T) {
					client := avngen.NewMockClient(t)
					client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
					client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
						Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithIP(tc.state, testIP)}, nil).Once()
					d := newPlan(t, testPSC)
					require.NoError(t, createView(t.Context(), client, d))
					require.Equal(t, testIP, d.Get("user_ip_address"))
				})
			}
		})

		t.Run("error", func(t *testing.T) {
			for _, tc := range []struct {
				name       string
				connection privatelink.ServicePrivatelinkGoogleConnectionListOut
				err        string
			}{
				{name: "different IP", connection: connectionWithIP(privatelink.ConnectionStateTypeActive, "10.0.0.3"), err: "endpoint IP cannot be set or changed"},
				{name: "GCP accepted without an IP", connection: connectionWithoutIP(privatelink.ConnectionStateTypeConnected), err: "endpoint IP cannot be set or changed"},
				{name: "invalid connected state with IP", connection: connectionWithIP(privatelink.ConnectionStateTypeConnected, testIP), err: "unexpected approval state"},
				{name: "unknown state", connection: connectionWithIP("unexpected", testIP), err: "unexpected approval state"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					client := avngen.NewMockClient(t)
					client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
					client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
						Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{tc.connection}, nil).Once()
					d := newPlan(t, testPSC)
					require.ErrorContains(t, createView(t.Context(), client, d), tc.err)
					require.Empty(t, d.ID())
				})
			}
		})
	})

	t.Run("API errors", func(t *testing.T) {
		t.Run("refresh", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			forbidden := avngen.Error{Status: http.StatusForbidden}
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(forbidden).Once()
			require.ErrorIs(t, createView(t.Context(), client, newPlan(t, testPSC)), forbidden)
		})

		t.Run("list", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			forbidden := avngen.Error{Status: http.StatusForbidden}
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(nil, forbidden).Once()
			require.ErrorIs(t, createView(t.Context(), client, newPlan(t, testPSC)), forbidden)
		})

		t.Run("approve", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			forbidden := avngen.Error{Status: http.StatusForbidden}
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).Return(nil, forbidden).Once()
			require.ErrorIs(t, createView(t.Context(), client, newPlan(t, testPSC)), forbidden)
		})
	})
}

func TestApprovalRetries(t *testing.T) {
	t.Run("failed response", func(t *testing.T) {
		t.Run("matching IP", func(t *testing.T) {
			for _, tc := range []struct {
				name string
				err  error
			}{
				{name: "conflict after commit", err: avngen.Error{Status: http.StatusConflict, Message: "Privatelink connection state is user-approved"}},
				{name: "timeout after commit", err: context.DeadlineExceeded},
				{name: "response lost after commit", err: io.ErrUnexpectedEOF},
				{name: "server error after commit", err: avngen.Error{Status: http.StatusBadGateway}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					synctest.Test(t, func(t *testing.T) {
						client := avngen.NewMockClient(t)
						client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
						client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
							Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
						client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
							Return(nil, tc.err).Once()
						client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
							Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithIP(privatelink.ConnectionStateTypeUserApproved, testIP)}, nil).Once()
						d := newPlan(t, testPSC)
						require.NoError(t, createView(t.Context(), client, d))
						require.Equal(t, "user-approved", d.Get("state"))
					})
				})
			}
		})

		t.Run("different IP", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				client := avngen.NewMockClient(t)
				client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
				client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
					Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
				client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
					Return(nil, avngen.Error{Status: http.StatusConflict}).Once()
				committed := connectionWithIP(privatelink.ConnectionStateTypeUserApproved, "10.0.0.3")
				client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
					Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{committed}, nil).Once()
				d := newPlan(t, testPSC)
				require.ErrorContains(t, createView(t.Context(), client, d), "endpoint IP cannot be set or changed")
			})
		})
	})

	t.Run("parent readiness", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Times(2)
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
				Return(nil, avngen.Error{Status: http.StatusConflict, Message: "Service is not active"}).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
				Return(approvedResponse(), nil).Once()
			require.NoError(t, createView(t.Context(), client, newPlan(t, testPSC)))
		})
	})

	t.Run("selection is pinned", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
				Return(nil, avngen.Error{Status: http.StatusNotFound}).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{{
					PrivatelinkConnectionId: "other", PscConnectionId: "other", State: privatelink.ConnectionStateTypePendingUserApproval,
				}}, nil)
			ctx, cancel := context.WithTimeout(t.Context(), 12*time.Second)
			defer cancel()
			err := createView(ctx, client, newPlanWithoutSelector(t))
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.ErrorContains(t, err, testPSC)
		})
	})
}

func TestApprovalDiscovery(t *testing.T) {
	t.Run("refresh throttling", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			var refreshed []time.Time
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).
				RunAndReturn(func(context.Context, string, string) error {
					refreshed = append(refreshed, time.Now())
					return nil
				}).Times(2)
			// Keep the target absent through the refresh deadline, then return it on the next poll.
			missingPolls := int(connectionRefreshInterval/connectionPollInterval) + 1
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{{PrivatelinkConnectionId: "other", PscConnectionId: "other"}}, nil).
				Times(missingPolls)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1", mock.Anything).
				Return(approvedResponse(), nil).Once()
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
			defer cancel()
			require.NoError(t, createView(ctx, client, newPlan(t, testPSC)))
			require.Len(t, refreshed, 2)
			require.GreaterOrEqual(t, refreshed[1].Sub(refreshed[0]), time.Minute)
			client.AssertNumberOfCalls(t, "ServicePrivatelinkGoogleConnectionList", missingPolls+1)
		})
	})

	t.Run("timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(nil, nil)
			ctx, cancel := context.WithTimeout(t.Context(), 12*time.Second)
			defer cancel()
			err := createView(ctx, client, newPlan(t, testPSC))
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.ErrorContains(t, err, testPSC)
		})
	})
}

func TestReadApproval(t *testing.T) {
	t.Run("selection", func(t *testing.T) {
		active := connectionWithIP(privatelink.ConnectionStateTypeActive, testIP)
		other := active
		other.PrivatelinkConnectionId, other.PscConnectionId = "plc-other", "other"

		t.Run("success", func(t *testing.T) {
			for _, tc := range []struct {
				name        string
				state       map[string]any
				connections []privatelink.ServicePrivatelinkGoogleConnectionListOut
			}{
				{name: "single connection import", connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{active}},
				{name: "selector", state: map[string]any{"psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}},
				{name: "legacy three-part ID", state: map[string]any{"id": testProject + "/" + testService + "/" + testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}},
				{name: "saved selector takes precedence over legacy ID", state: map[string]any{"id": testProject + "/" + testService + "/other", "psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}},
				{name: "saved internal ID", state: map[string]any{"privatelink_connection_id": "plc1"}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}},
				{name: "rediscovered PSC endpoint", state: map[string]any{"privatelink_connection_id": "plc-previous", "psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					prior := map[string]any{"id": testProject + "/" + testService}
					for key, value := range tc.state {
						prior[key] = value
					}
					d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestState(prior))
					require.NoError(t, err)
					client := avngen.NewMockClient(t)
					client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(tc.connections, nil).Once()
					require.NoError(t, readView(t.Context(), client, d))
					require.Equal(t, testProject+"/"+testService, d.ID())
					require.Equal(t, testProject, d.Get("project"))
					require.Equal(t, testService, d.Get("service_name"))
					require.Equal(t, "plc1", d.Get("privatelink_connection_id"))
					require.Equal(t, testPSC, d.Get("psc_connection_id"))
					require.Equal(t, testIP, d.Get("user_ip_address"))
					require.Equal(t, "active", d.Get("state"))
				})
			}
		})

		t.Run("error", func(t *testing.T) {
			withoutPSC := other
			withoutPSC.PscConnectionId = ""
			for _, tc := range []struct {
				name        string
				state       map[string]any
				connections []privatelink.ServicePrivatelinkGoogleConnectionListOut
				err         string
			}{
				{name: "ambiguous import", connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other, active}, err: "set psc_connection_id"},
				{name: "duplicate PSC IDs", state: map[string]any{"psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{active, active}, err: "multiple privatelink connections match"},
				{name: "selected endpoint missing", state: map[string]any{"psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other}, err: "not found"},
				{name: "missing internal ID must not select another endpoint", state: map[string]any{"privatelink_connection_id": "plc-gone"}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{other}, err: "not found"},
				{name: "empty PSC ID cannot match a selector", state: map[string]any{"psc_connection_id": testPSC}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{withoutPSC}, err: "not found"},
				{name: "empty PSC IDs are not fallback keys", state: map[string]any{"privatelink_connection_id": "plc-gone", "psc_connection_id": ""}, connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{withoutPSC}, err: "not found"},
				{name: "no connections", err: "not found"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					prior := map[string]any{"id": testProject + "/" + testService}
					for key, value := range tc.state {
						prior[key] = value
					}
					d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestState(prior))
					require.NoError(t, err)
					client := avngen.NewMockClient(t)
					client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(tc.connections, nil).Once()
					require.ErrorContains(t, readView(t.Context(), client, d), tc.err)
				})
			}
		})
	})

	t.Run("null IP over HTTP", func(t *testing.T) {
		for _, connectionState := range []string{"pending-user-approval", "connected"} {
			t.Run(connectionState, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/v1/project/test-project/service/test-service/privatelink/google/connections", r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"connections":[{"privatelink_connection_id":"plc1","psc_connection_id":"50079288774361602","state":"`+connectionState+`","user_ip_address":null}]}`)
				}))
				defer server.Close()
				client, err := avngen.NewClient(avngen.TokenOpt("test"), avngen.HostOpt(server.URL), avngen.DoerOpt(server.Client()), avngen.DebugOpt(false))
				require.NoError(t, err)
				state := legacyState(t, map[string]any{
					"id": testProject + "/" + testService, "privatelink_connection_id": "plc1", "psc_connection_id": testPSC, "user_ip_address": testIP,
				})
				r := configuredResource(t, client)
				response := resource.ReadResponse{State: state}
				r.Read(t.Context(), resource.ReadRequest{State: state}, &response)
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				var ip types.String
				require.False(t, response.State.GetAttribute(t.Context(), path.Root("user_ip_address"), &ip).HasError())
				require.True(t, ip.IsNull(), "an unapproved endpoint must not retain its previous IP")
				var gotState string
				require.False(t, response.State.GetAttribute(t.Context(), path.Root("state"), &gotState).HasError())
				require.Equal(t, connectionState, gotState)
			})
		}
	})

	t.Run("missing or forbidden", func(t *testing.T) {
		t.Run("missing connection", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(nil, nil).Once()
			state := legacyState(t, map[string]any{"id": testProject + "/" + testService, "privatelink_connection_id": "plc1", "psc_connection_id": testPSC})
			response := resource.ReadResponse{State: state}
			configuredResource(t, client).Read(t.Context(), resource.ReadRequest{State: state}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			require.True(t, response.State.Raw.IsNull())
		})

		t.Run("missing parent", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return(nil, avngen.Error{Status: http.StatusNotFound}).Once()
			state := legacyState(t, map[string]any{"id": testProject + "/" + testService, "privatelink_connection_id": "plc1", "psc_connection_id": testPSC})
			response := resource.ReadResponse{State: state}
			configuredResource(t, client).Read(t.Context(), resource.ReadRequest{State: state}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			require.True(t, response.State.Raw.IsNull())
		})

		t.Run("forbidden", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return(nil, avngen.Error{Status: http.StatusForbidden}).Once()
			state := legacyState(t, map[string]any{"id": testProject + "/" + testService, "privatelink_connection_id": "plc1", "psc_connection_id": testPSC})
			response := resource.ReadResponse{State: state}
			configuredResource(t, client).Read(t.Context(), resource.ReadRequest{State: state}, &response)
			require.True(t, response.Diagnostics.HasError())
			require.False(t, response.State.Raw.IsNull())
		})
	})
}

func TestUpdateApproval(t *testing.T) {
	t.Run("imported pending connection", func(t *testing.T) {
		prior := map[string]any{
			"id": testProject + "/" + testService, "project": testProject, "service_name": testService,
			"privatelink_connection_id": "plc1", "psc_connection_id": testPSC, "user_ip_address": "",
		}
		d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields,
			adapter.WithTestState(prior), adapter.WithTestPlan(map[string]any{
				"project": testProject, "service_name": testService, "user_ip_address": testIP,
			}))
		require.NoError(t, err)
		client := avngen.NewMockClient(t)
		client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
		client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
			Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{
				connectionWithoutIP(privatelink.ConnectionStateTypePendingUserApproval),
				{PrivatelinkConnectionId: "other", PscConnectionId: "other"},
			}, nil).Once()
		client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1",
			&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: testIP}).Return(approvedResponse(), nil).Once()
		require.NoError(t, updateView(t.Context(), client, d))
		require.Equal(t, testIP, d.Get("user_ip_address"))
	})

	t.Run("immutable approved IP", func(t *testing.T) {
		prior := map[string]any{
			"id": testProject + "/" + testService, "project": testProject, "service_name": testService,
			"privatelink_connection_id": "plc1", "psc_connection_id": testPSC, "user_ip_address": "10.0.0.3",
		}
		d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields,
			adapter.WithTestState(prior), adapter.WithTestPlan(map[string]any{
				"project": testProject, "service_name": testService, "user_ip_address": testIP,
			}))
		require.NoError(t, err)
		conn := connectionWithIP(privatelink.ConnectionStateTypeActive, "10.0.0.3")
		client := avngen.NewMockClient(t)
		client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Once()
		client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
			Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{conn, {PrivatelinkConnectionId: "other", PscConnectionId: "other"}}, nil).Once()
		require.ErrorContains(t, updateView(t.Context(), client, d), "endpoint IP cannot be set or changed")
	})

	t.Run("timeouts only", func(t *testing.T) {
		d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields,
			adapter.WithTestState(map[string]any{"id": testProject + "/" + testService, "user_ip_address": testIP}),
			adapter.WithTestPlan(map[string]any{"user_ip_address": testIP, "timeouts": map[string]any{"read": "1m"}}))
		require.NoError(t, err)
		require.ErrorIs(t, updateView(t.Context(), avngen.NewMockClient(t), d), adapter.ErrNoChanges)
	})
}

func TestApprovalImport(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		for _, id := range []string{testProject + "/" + testService, testProject + "/" + testService + "/" + testPSC} {
			t.Run(id, func(t *testing.T) {
				client := avngen.NewMockClient(t)
				r := configuredResource(t, client)
				sch := resourceSchema(t.Context())
				response := resource.ImportStateResponse{State: tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(t.Context()), nil)}}
				r.(resource.ResourceWithImportState).ImportState(t.Context(), resource.ImportStateRequest{ID: id}, &response)
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				var normalized string
				require.False(t, response.State.GetAttribute(t.Context(), path.Root("id"), &normalized).HasError())
				require.Equal(t, testProject+"/"+testService, normalized)
				client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
					Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{connectionWithIP(privatelink.ConnectionStateTypeActive, testIP)}, nil).Once()
				read := resource.ReadResponse{State: response.State}
				r.Read(t.Context(), resource.ReadRequest{State: response.State}, &read)
				require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
				var psc string
				require.False(t, read.State.GetAttribute(t.Context(), path.Root("psc_connection_id"), &psc).HasError())
				require.Equal(t, testPSC, psc)
			})
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		for _, id := range []string{"", "project", "project/", "/service", "project/service/", "project/service/psc/extra"} {
			t.Run(id, func(t *testing.T) {
				_, _, _, err := parseApprovalID(id)
				require.Error(t, err)
			})
		}
	})
}

func legacyState(t *testing.T, prior map[string]any) tfsdk.State {
	t.Helper()
	raw, err := json.Marshal(prior)
	require.NoError(t, err)
	state := tfsdk.State{Schema: resourceSchema(t.Context())}
	state.Raw, err = (&tfprotov6.RawState{JSON: raw}).Unmarshal(state.Schema.Type().TerraformType(t.Context()))
	require.NoError(t, err)
	return state
}

type testProviderData struct{ client avngen.Client }

func (testProviderData) GetClient() *aiven.Client      { return nil }
func (p testProviderData) GetGenClient() avngen.Client { return p.client }

func configuredResource(t *testing.T, client avngen.Client) resource.Resource {
	t.Helper()
	r := adapter.NewResource(ResourceOptions)
	response := resource.ConfigureResponse{}
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: testProviderData{client}}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	return r
}
