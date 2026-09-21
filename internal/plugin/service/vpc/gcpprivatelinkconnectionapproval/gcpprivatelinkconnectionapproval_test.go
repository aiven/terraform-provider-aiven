package gcpprivatelinkconnectionapproval

import (
	"context"
	"errors"
	"io"
	"maps"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testProject = "test-project"
	testService = "test-service"
	testIP      = "10.0.0.2"
	testID      = testProject + "/" + testService
)

type connection = privatelink.ServicePrivatelinkGoogleConnectionListOut

func testConnection(state privatelink.ConnectionStateType) connection {
	c := connection{PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", State: state}
	if state == privatelink.ConnectionStateTypeUserApproved || state == privatelink.ConnectionStateTypeActive {
		c.UserIPAddress = testIP
	}
	return c
}

func testData(t *testing.T, opts ...adapter.ResourceDataOpt) adapter.ResourceData {
	t.Helper()
	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), opts...)
	require.NoError(t, err)
	return d
}

func testPlan() map[string]any {
	return map[string]any{"project": testProject, "service_name": testService, "user_ip_address": testIP}
}

func TestApproveConnection(t *testing.T) {
	for _, state := range []privatelink.ConnectionStateType{
		privatelink.ConnectionStateTypePendingUserApproval,
		privatelink.ConnectionStateTypeUserApproved,
		privatelink.ConnectionStateTypeActive,
		privatelink.ConnectionStateTypeConnected,
	} {
		t.Run(string(state), func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
				Return([]connection{testConnection(state)}, nil).Once()
			if state == privatelink.ConnectionStateTypePendingUserApproval {
				client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(t.Context(), testProject, testService, "plc1",
					&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: testIP}).
					Return(&privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
						PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", UserIPAddress: testIP,
						State: privatelink.ServicePrivatelinkGoogleConnectionStateTypeUserApproved,
					}, nil).Once()
			}
			d := testData(t, adapter.WithTestPlan(testPlan()))
			err := createView(t.Context(), client, d)
			if state == privatelink.ConnectionStateTypeConnected {
				require.ErrorContains(t, err, ipChangeError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testID, d.ID())
			require.Equal(t, "plc1", d.Get("privatelink_connection_id"))
			require.Equal(t, "psc1", d.Get("psc_connection_id"))
			require.Equal(t, testIP, d.Get("user_ip_address"))
		})
	}
}

func TestApproveConnectionRecoversAfterError(t *testing.T) {
	for name, approveErr := range map[string]error{
		"conflict":      avngen.Error{Status: http.StatusConflict, Message: "Privatelink connection state is user-approved"},
		"lost response": io.EOF,
	} {
		t.Run(name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
				Return([]connection{testConnection(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(t.Context(), testProject, testService, "plc1", mock.Anything).
				Return(nil, approveErr).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
				Return([]connection{testConnection(privatelink.ConnectionStateTypeUserApproved)}, nil).Once()
			d := testData(t, adapter.WithTestPlan(testPlan()))
			require.NoError(t, createView(t.Context(), client, d))
			require.Equal(t, "user-approved", d.Get("state"))
			require.Equal(t, testIP, d.Get("user_ip_address"))
		})
	}
}

func TestApproveConnectionRetriesOnlyAfterReadingPending(t *testing.T) {
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Times(2)
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
		Return([]connection{testConnection(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Times(3)
	client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(t.Context(), testProject, testService, "plc1", mock.Anything).
		Return(nil, avngen.Error{Status: http.StatusConflict, Message: "Privatelink state is creating"}).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(t.Context(), testProject, testService, "plc1", mock.Anything).
		Return(&privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
			PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", UserIPAddress: testIP,
			State: privatelink.ServicePrivatelinkGoogleConnectionStateTypeUserApproved,
		}, nil).Once()
	require.NoError(t, createView(t.Context(), client, testData(t, adapter.WithTestPlan(testPlan()))))
}

func TestApproveConnectionDiscovery(t *testing.T) {
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Times(3)
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).Return(nil, nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
		Return([]connection{{PrivatelinkConnectionId: "other", PscConnectionId: "other"}}, nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
		Return([]connection{testConnection(privatelink.ConnectionStateTypeActive), {PrivatelinkConnectionId: "other", PscConnectionId: "other"}}, nil).Once()
	plan := testPlan()
	plan["psc_connection_id"] = "psc1"
	d := testData(t, adapter.WithTestPlan(plan))
	require.NoError(t, createView(t.Context(), client, d))
	require.Equal(t, "plc1", d.Get("privatelink_connection_id"))
}

func TestApproveConnectionErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		refreshErr error
		listErr    error
		approveErr error
		after      []connection
		want       string
	}{
		"refresh error":   {refreshErr: errors.New("refresh failed"), want: "refresh failed"},
		"list error":      {listErr: errors.New("list failed"), want: "list failed"},
		"invalid request": {approveErr: avngen.Error{Status: http.StatusBadRequest, Message: "invalid IP"}, after: []connection{testConnection(privatelink.ConnectionStateTypePendingUserApproval)}, want: "invalid IP"},
		"disappeared":     {approveErr: io.EOF, want: "connection not found"},
		"different IP":    {approveErr: io.EOF, after: []connection{{PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.3"}}, want: ipChangeError},
	} {
		t.Run(name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(tc.refreshErr).Once()
			if tc.refreshErr == nil {
				client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
					Return([]connection{testConnection(privatelink.ConnectionStateTypePendingUserApproval)}, tc.listErr).Once()
			}
			if tc.approveErr != nil {
				client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(t.Context(), testProject, testService, "plc1", mock.Anything).Return(nil, tc.approveErr).Once()
				client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).Return(tc.after, nil).Once()
			}
			require.ErrorContains(t, createView(t.Context(), client, testData(t, adapter.WithTestPlan(testPlan()))), tc.want)
		})
	}
}

func TestApproveConnectionCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(ctx, testProject, testService).Return(nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(ctx, testProject, testService).
		Run(func(context.Context, string, string) { cancel() }).Return(nil, nil).Once()
	err := createView(ctx, client, testData(t, adapter.WithTestPlan(testPlan())))
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, adapter.ErrRefreshStateDesired)
}

func TestExplicitEmptySelector(t *testing.T) {
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).
		Return([]connection{testConnection(privatelink.ConnectionStateTypeActive)}, nil).Twice()
	plan := testPlan()
	plan["psc_connection_id"] = ""
	d := testData(t, adapter.WithTestPlan(plan))
	require.NoError(t, createView(t.Context(), client, d))
	// Both nil and "" satisfy Empty, but only "" preserves the configured value.
	require.IsType(t, "", d.Get("psc_connection_id"))
	require.Empty(t, d.Get("psc_connection_id"))
	require.Equal(t, "plc1", d.Get("privatelink_connection_id"))

	state := testData(t, adapter.WithTestState(maps.Clone(plan)))
	require.NoError(t, readView(t.Context(), client, state))
	require.IsType(t, "", state.Get("psc_connection_id"))
	require.Empty(t, state.Get("psc_connection_id"))
	require.Equal(t, "plc1", state.Get("privatelink_connection_id"))
}

func TestReadDuringApplyPreservesConfiguredIP(t *testing.T) {
	for name, tc := range map[string]struct {
		conn    connection
		pending bool
	}{
		"pending response":      {conn: testConnection(privatelink.ConnectionStateTypePendingUserApproval), pending: true},
		"accepted without IP":   {conn: testConnection(privatelink.ConnectionStateTypeConnected)},
		"different approved IP": {conn: connection{PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.3"}},
	} {
		t.Run(name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).Return([]connection{tc.conn}, nil).Once()
			plan := testPlan()
			plan["id"], plan["privatelink_connection_id"] = testID, "plc1"
			d := testData(t, adapter.WithTestPlan(plan), adapter.WithConfig(tfsdk.Config{Raw: testTFValue(t, testPlan())}))
			err := readView(t.Context(), client, d)
			if tc.pending {
				require.ErrorIs(t, err, adapter.ErrRefreshStateDesired)
			} else {
				require.ErrorContains(t, err, ipChangeError)
			}
			require.Equal(t, testIP, d.Get("user_ip_address"))
		})
	}
}

func TestApprovedConnectionDisappearanceIsTerminal(t *testing.T) {
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(t.Context(), testProject, testService).Return(nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).Return(nil, nil).Once()
	plan := testPlan()
	plan["id"], plan["privatelink_connection_id"] = testID, "plc1"
	d := testData(t, adapter.WithTestPlan(plan), adapter.WithConfig(tfsdk.Config{Raw: testTFValue(t, testPlan())}))
	err := readView(t.Context(), client, d)
	require.ErrorIs(t, err, adapter.ErrRefreshStateFailed)
	require.False(t, adapter.IsNotFound(err), "the post-create poller must not retry a deleted connection")
}

func TestReadConnectionSelection(t *testing.T) {
	one := testConnection(privatelink.ConnectionStateTypeActive)
	other := connection{PrivatelinkConnectionId: "plc2", PscConnectionId: "psc2", State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.3"}
	for name, tc := range map[string]struct {
		id           string
		selector     string
		connectionID string
		connections  []connection
		wantID       string
		wantErr      error
	}{
		"single connection":                     {id: testID, connections: []connection{one}, wantID: "plc1"},
		"PSC selector":                          {id: testID, selector: "psc2", connections: []connection{one, other}, wantID: "plc2"},
		"three part import":                     {id: testID + "/psc2", connections: []connection{one, other}, wantID: "plc2"},
		"existing selector takes precedence":    {id: testID + "/psc2", selector: "psc1", connections: []connection{one, other}, wantID: "plc1"},
		"stored internal ID":                    {id: testID, connectionID: "plc2", connections: []connection{one, other}, wantID: "plc2"},
		"missing internal ID is not retargeted": {id: testID, connectionID: "gone", connections: []connection{one}, wantErr: adapter.ErrNotFound},
		"empty list":                            {id: testID, wantErr: adapter.ErrNotFound},
		"missing PSC":                           {id: testID, selector: "missing", connections: []connection{one}, wantErr: adapter.ErrNotFound},
		"ambiguous without selector":            {id: testID, connections: []connection{one, other}, wantErr: adapter.ErrMultiple},
		"duplicate PSC":                         {id: testID, selector: "psc1", connections: []connection{one, one}, wantErr: adapter.ErrMultiple},
	} {
		t.Run(name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(t.Context(), testProject, testService).Return(tc.connections, nil).Once()
			d := testData(t, adapter.WithTestState(map[string]any{"id": tc.id, "psc_connection_id": tc.selector, "privatelink_connection_id": tc.connectionID}))
			err := readView(t.Context(), client, d)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testID, d.ID())
			require.Equal(t, testProject, d.Get("project"))
			require.Equal(t, testService, d.Get("service_name"))
			require.Equal(t, tc.wantID, d.Get("privatelink_connection_id"))
		})
	}
}

func TestModifyPlan(t *testing.T) {
	for name, tc := range map[string]struct {
		plan        map[string]any
		state       map[string]any
		newResource bool
		wantErr     bool
	}{
		"unchanged":                   {},
		"IP change":                   {plan: map[string]any{"user_ip_address": "10.0.0.3"}, wantErr: true},
		"unknown IP":                  {plan: map[string]any{"user_ip_address": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}},
		"new endpoint":                {plan: map[string]any{"user_ip_address": "10.0.0.3", "psc_connection_id": "psc2"}},
		"unknown endpoint":            {plan: map[string]any{"user_ip_address": "10.0.0.3", "psc_connection_id": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}},
		"new project":                 {plan: map[string]any{"user_ip_address": "10.0.0.3", "project": "other"}},
		"new service":                 {plan: map[string]any{"user_ip_address": "10.0.0.3", "service_name": "other"}},
		"imported pending connection": {state: map[string]any{"state": "pending-user-approval", "user_ip_address": ""}},
		"connected without IP":        {state: map[string]any{"state": "connected", "user_ip_address": ""}, wantErr: true},
		"new resource":                {newResource: true},
	} {
		t.Run(name, func(t *testing.T) {
			plan := testPlan()
			plan["psc_connection_id"] = "psc1"
			state := maps.Clone(plan)
			state["id"], state["state"] = testID, "active"
			maps.Copy(plan, tc.plan)
			maps.Copy(state, tc.state)
			opts := []adapter.ResourceDataOpt{adapter.WithTestPlan(plan)}
			if !tc.newResource {
				opts = append(opts, adapter.WithTestState(state))
			}
			err := modifyPlan(t.Context(), nil, testData(t, opts...))
			if tc.wantErr {
				require.ErrorContains(t, err, ipChangeError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateWaitsForActiveAndPreservesApprovedState(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(map[bool]string{false: "approved but not active", true: "active"}[active], func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Times(2)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]connection{testConnection(privatelink.ConnectionStateTypeUserApproved)}, nil).Once()
			state := privatelink.ConnectionStateTypeUserApproved
			if active {
				state = privatelink.ConnectionStateTypeActive
			}
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Run(func(context.Context, string, string) {
					if !active {
						// Stop before active to check that a failed wait retains the approved ID.
						cancel()
					}
				}).
				Return([]connection{testConnection(state)}, nil).Once()
			opts := ResourceOptions
			opts.Create = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
				return createView(ctx, client, d)
			}
			opts.Read = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
				return readView(ctx, client, d)
			}
			raw := testTFValue(t, testPlan())
			rsp := resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema(ctx), Raw: tftypes.NewValue(raw.Type(), nil)}}
			adapter.NewResource(opts).Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Raw: raw}, Config: tfsdk.Config{Raw: raw}}, &rsp)
			require.Equal(t, !active, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
			var id, gotState string
			require.False(t, rsp.State.GetAttribute(t.Context(), path.Root("id"), &id).HasError())
			require.False(t, rsp.State.GetAttribute(t.Context(), path.Root("state"), &gotState).HasError())
			require.Equal(t, testID, id)
			require.Equal(t, string(state), gotState)
		})
	}
}

func TestImportAndMissingConnection(t *testing.T) {
	for _, id := range []string{testID, testID + "/psc1"} {
		t.Run(id, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
				Return([]connection{testConnection(privatelink.ConnectionStateTypeActive)}, nil).Once()
			opts := ResourceOptions
			opts.Read = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
				return readView(ctx, client, d)
			}
			r := adapter.NewResource(opts)
			importRsp := resource.ImportStateResponse{State: tfsdk.State{Schema: resourceSchema(t.Context()), Raw: testTFValue(t, nil)}}
			r.(resource.ResourceWithImportState).ImportState(t.Context(), resource.ImportStateRequest{ID: id}, &importRsp)
			require.False(t, importRsp.Diagnostics.HasError(), "%v", importRsp.Diagnostics)
			readRsp := resource.ReadResponse{State: importRsp.State}
			r.Read(t.Context(), resource.ReadRequest{State: importRsp.State}, &readRsp)
			require.False(t, readRsp.Diagnostics.HasError(), "%v", readRsp.Diagnostics)
			var normalizedID string
			require.False(t, readRsp.State.GetAttribute(t.Context(), path.Root("id"), &normalizedID).HasError())
			require.Equal(t, testID, normalizedID)
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).Return(nil, nil).Once()
			missingRsp := resource.ReadResponse{State: readRsp.State}
			r.Read(t.Context(), resource.ReadRequest{State: readRsp.State}, &missingRsp)
			require.False(t, missingRsp.Diagnostics.HasError(), "%v", missingRsp.Diagnostics)
			require.True(t, missingRsp.State.Raw.IsNull())
		})
	}
}

func TestUpdateApprovesImportedPendingConnection(t *testing.T) {
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, testProject, testService).Return(nil).Twice()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
		Return([]connection{testConnection(privatelink.ConnectionStateTypePendingUserApproval)}, nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, testProject, testService, "plc1",
		&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: testIP}).
		Return(&privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
			PrivatelinkConnectionId: "plc1", PscConnectionId: "psc1", UserIPAddress: testIP,
			State: privatelink.ServicePrivatelinkGoogleConnectionStateTypeUserApproved,
		}, nil).Once()
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, testProject, testService).
		Return([]connection{testConnection(privatelink.ConnectionStateTypeActive)}, nil).Once()
	opts := ResourceOptions
	update := opts.Update
	opts.Update = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
		return update(ctx, client, d)
	}
	opts.Read = func(ctx context.Context, _ avngen.Client, d adapter.ResourceData) error {
		return readView(ctx, client, d)
	}
	plan := testPlan()
	plan["id"], plan["psc_connection_id"], plan["privatelink_connection_id"] = testID, "psc1", "plc1"
	prior := maps.Clone(plan)
	prior["user_ip_address"], prior["state"] = "", "pending-user-approval"
	state := tfsdk.State{Schema: resourceSchema(t.Context()), Raw: testTFValue(t, prior)}
	rsp := resource.UpdateResponse{State: state}
	adapter.NewResource(opts).Update(t.Context(), resource.UpdateRequest{
		Plan: tfsdk.Plan{Raw: testTFValue(t, plan)}, Config: tfsdk.Config{Raw: testTFValue(t, testPlan())}, State: state,
	}, &rsp)
	require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
	var ip, status string
	require.False(t, rsp.State.GetAttribute(t.Context(), path.Root("user_ip_address"), &ip).HasError())
	require.False(t, rsp.State.GetAttribute(t.Context(), path.Root("state"), &status).HasError())
	require.Equal(t, testIP, ip)
	require.Equal(t, "active", status)
}

func testTFValue(t *testing.T, values map[string]any) tftypes.Value {
	t.Helper()
	typ := resourceSchema(t.Context()).Type().TerraformType(t.Context()).(tftypes.Object)
	attrs := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for key, attrType := range typ.AttributeTypes {
		attrs[key] = tftypes.NewValue(attrType, values[key])
	}
	return tftypes.NewValue(typ, attrs)
}
