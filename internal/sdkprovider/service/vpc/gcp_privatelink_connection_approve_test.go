package vpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

const (
	testProject = "test-project"
	testService = "test-service"
)

func useInstantStateChangeTimings(t *testing.T) {
	t.Helper()

	testMu.Lock()
	originalDelay := gcpPSCApprovalStateChangeDelay
	originalMinTimeout := gcpPSCApprovalStateChangeMinTimeout
	gcpPSCApprovalStateChangeDelay = 0
	gcpPSCApprovalStateChangeMinTimeout = 0
	testMu.Unlock()

	t.Cleanup(func() {
		testMu.Lock()
		gcpPSCApprovalStateChangeDelay = originalDelay
		gcpPSCApprovalStateChangeMinTimeout = originalMinTimeout
		testMu.Unlock()
	})
}

func newApprovalResourceData(t *testing.T, userIPAddress string) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
		"project":         testProject,
		"service_name":    testService,
		"user_ip_address": userIPAddress,
	})
}

func newApprovalResourceDataWithPSCConnectionID(t *testing.T, userIPAddress, pscConnectionID string) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
		"project":           testProject,
		"service_name":      testService,
		"user_ip_address":   userIPAddress,
		"psc_connection_id": pscConnectionID,
	})
}

type fakeGCPPrivatelink struct {
	project string
	service string

	refreshErr         error
	connectionsListErr error

	connections   []aiven.GCPPrivatelinkConnectionResponse
	approveCalls  []string
	approveUserIP []string

	connectionsListCalls                int
	connectionsListSequence             [][]aiven.GCPPrivatelinkConnectionResponse
	emptyConnectionsListCallsRemaining  int
	returnedEmptyConnectionsListAtLeast bool
}

func newFakeGCPPrivatelink(
	initialConnections []aiven.GCPPrivatelinkConnectionResponse,
) *fakeGCPPrivatelink {
	return &fakeGCPPrivatelink{
		project:     testProject,
		service:     testService,
		connections: append([]aiven.GCPPrivatelinkConnectionResponse(nil), initialConnections...),
	}
}

func (f *fakeGCPPrivatelink) assertProjectService(project, service string) error {
	if project != f.project || service != f.service {
		return fmt.Errorf("unexpected project/service: %s/%s", project, service)
	}
	return nil
}

func (f *fakeGCPPrivatelink) AddConnection(conn aiven.GCPPrivatelinkConnectionResponse) {
	f.connections = append(f.connections, conn)
}

func (f *fakeGCPPrivatelink) Refresh(_ context.Context, project, serviceName string) error {
	if f.refreshErr != nil {
		return f.refreshErr
	}

	return f.assertProjectService(project, serviceName)
}

func (f *fakeGCPPrivatelink) ConnectionsList(
	_ context.Context,
	project,
	serviceName string,
) (*aiven.GCPPrivatelinkConnectionsResponse, error) {
	if f.connectionsListErr != nil {
		return nil, f.connectionsListErr
	}

	if err := f.assertProjectService(project, serviceName); err != nil {
		return nil, err
	}

	f.connectionsListCalls++
	if f.emptyConnectionsListCallsRemaining > 0 {
		f.emptyConnectionsListCallsRemaining--
		f.returnedEmptyConnectionsListAtLeast = true
		return &aiven.GCPPrivatelinkConnectionsResponse{Connections: nil}, nil
	}

	if seq := f.connectionsListSequence; len(seq) > 0 && f.connectionsListCalls <= len(seq) {
		return &aiven.GCPPrivatelinkConnectionsResponse{
			Connections: append([]aiven.GCPPrivatelinkConnectionResponse(nil), seq[f.connectionsListCalls-1]...),
		}, nil
	}

	return &aiven.GCPPrivatelinkConnectionsResponse{
		Connections: append([]aiven.GCPPrivatelinkConnectionResponse(nil), f.connections...),
	}, nil
}

func (f *fakeGCPPrivatelink) ConnectionApprove(
	_ context.Context,
	project, serviceName, connID string,
	req aiven.GCPPrivatelinkConnectionApproveRequest,
) error {
	if err := f.assertProjectService(project, serviceName); err != nil {
		return err
	}

	f.approveCalls = append(f.approveCalls, connID)
	f.approveUserIP = append(f.approveUserIP, req.UserIPAddress)

	for i := range f.connections {
		if f.connections[i].PrivatelinkConnectionID == connID {
			f.connections[i].UserIPAddress = req.UserIPAddress
			f.connections[i].State = "connected"
			return nil
		}
	}

	return fmt.Errorf("connection not found: %s", connID)
}

func (f *fakeGCPPrivatelink) ConnectionGet(
	_ context.Context,
	project, serviceName string,
	connID *string,
) (*aiven.GCPPrivatelinkConnectionResponse, error) {
	if err := f.assertProjectService(project, serviceName); err != nil {
		return nil, err
	}

	if connID == nil || *connID == "" {
		return nil, fmt.Errorf("connection id is required")
	}

	for _, it := range f.connections {
		if it.PrivatelinkConnectionID == *connID {
			conn := it
			return &conn, nil
		}
	}

	return nil, aiven.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("connection not found: %s", *connID)}
}

func TestGCPPrivatelinkConnectionApprovalUpdate(t *testing.T) {
	useInstantStateChangeTimings(t)

	t.Run("Approves when only one connection is pending-user-approval", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "pending-user-approval",
			},
		})

		d := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.False(t, diags.HasError())

		require.Equal(t, testProject+"/"+testService, d.Id())
		require.Equal(t, "plc1", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc1", d.Get("psc_connection_id").(string))
		require.Equal(t, "connected", d.Get("state").(string))
		require.Equal(t, "10.0.0.2", d.Get("user_ip_address").(string))
		require.Equal(t, []string{"plc1"}, fake.approveCalls)
		require.Equal(t, []string{"10.0.0.2"}, fake.approveUserIP)
	})

	t.Run("Doesn't approve when already active", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "active",
				UserIPAddress:           "10.0.0.2",
			},
		})

		d := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.False(t, diags.HasError())
		require.Empty(t, fake.approveCalls)
		require.Empty(t, fake.approveUserIP)
		require.Equal(t, testProject+"/"+testService, d.Id())
		require.Equal(t, "plc1", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc1", d.Get("psc_connection_id").(string))
		require.Equal(t, "active", d.Get("state").(string))
		require.Equal(t, "10.0.0.2", d.Get("user_ip_address").(string))
	})

	t.Run("Waits until a connection exists", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "pending-user-approval",
			},
		})
		fake.emptyConnectionsListCallsRemaining = 1

		d := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.False(t, diags.HasError())
		require.True(t, fake.returnedEmptyConnectionsListAtLeast)
		require.GreaterOrEqual(t, fake.connectionsListCalls, 2)
		require.Equal(t, []string{"plc1"}, fake.approveCalls)
	})

	t.Run("Errors when multiple connections exist without PSC selector", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "connected", UserIPAddress: "10.0.0.2"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "pending-user-approval"},
		})

		d := newApprovalResourceData(t, "10.0.0.3")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections found")
	})

	t.Run("Errors when another connection exists without PSC selector", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "pending-user-approval",
			},
		})

		d1 := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d1, fake)
		require.False(t, diags.HasError())

		fake.AddConnection(aiven.GCPPrivatelinkConnectionResponse{
			PrivatelinkConnectionID: "plc2",
			PSCConnectionID:         "psc2",
			State:                   "pending-user-approval",
		})

		d2 := newApprovalResourceData(t, "10.0.0.3")

		diags2 := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d2, fake)
		require.True(t, diags2.HasError(), "expected error, got: %#v", diags2)
		require.Contains(t, diags2[0].Summary, "multiple privatelink connections found")
	})

	t.Run("Approves selected by PSC connection ID when multiple connections exist", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "connected", UserIPAddress: "10.0.0.2"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "pending-user-approval"},
		})

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.3", "psc2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.False(t, diags.HasError())
		require.Equal(t, testProject+"/"+testService, d.Id())
		require.Equal(t, "plc2", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc2", d.Get("psc_connection_id").(string))
		require.Equal(t, []string{"plc2"}, fake.approveCalls)
		require.Equal(t, []string{"10.0.0.3"}, fake.approveUserIP)
	})

	t.Run("Errors when PSC connection ID matches multiple connections", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "pending-user-approval"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc1", State: "pending-user-approval"},
		})

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.3", "psc1")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections match psc_connection_id")
	})

	t.Run("Errors when PSC connection ID isn't found (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: make it succeed (target includes pending-user-approval).
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc2", State: "pending-user-approval"},
			},
			// selection in Update: make it fail (psc2 disappeared).
			{
				{PrivatelinkConnectionID: "plc-other", PSCConnectionID: "psc1", State: "pending-user-approval"},
			},
		}

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.2", "psc2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "psc_connection_id \"psc2\" not found")
	})

	t.Run("Errors when PSC connection ID isn't found among multiple connections (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: single match so it succeeds.
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc2", State: "pending-user-approval"},
			},
			// selection in Update: multiple connections, but none match psc2.
			{
				{PrivatelinkConnectionID: "plc-1", PSCConnectionID: "psc1", State: "pending-user-approval"},
				{PrivatelinkConnectionID: "plc-2", PSCConnectionID: "psc3", State: "pending-user-approval"},
			},
		}

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.2", "psc2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "psc_connection_id \"psc2\" not found")
	})

	t.Run("Errors when PSC connection ID isn't found because connections list becomes empty (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: single match so it succeeds.
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc2", State: "pending-user-approval"},
			},
			// selection in Update: now it's empty.
			nil,
		}

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.2", "psc2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "psc_connection_id \"psc2\" not found")
	})

	t.Run("Errors when PSC connection ID matches multiple connections (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: single match so it doesn't error.
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc1", State: "pending-user-approval"},
			},
			// selection in Update: now it's ambiguous.
			{
				{PrivatelinkConnectionID: "plc-1", PSCConnectionID: "psc1", State: "pending-user-approval"},
				{PrivatelinkConnectionID: "plc-2", PSCConnectionID: "psc1", State: "pending-user-approval"},
			},
		}

		d := newApprovalResourceDataWithPSCConnectionID(t, "10.0.0.2", "psc1")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections match psc_connection_id")
	})

	t.Run("Errors when connections list becomes empty without PSC selector (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: single connection so it succeeds.
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc1", State: "pending-user-approval"},
			},
			// selection in Update: now it's empty.
			nil,
		}

		d := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "no privatelink connections found")
	})

	t.Run("Errors when multiple connections exist without PSC selector (selection step)", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListSequence = [][]aiven.GCPPrivatelinkConnectionResponse{
			// waitForGCPConnectionState: single connection so it doesn't error.
			{
				{PrivatelinkConnectionID: "plc-ok", PSCConnectionID: "psc1", State: "pending-user-approval"},
			},
			// selection in Update: now there are two.
			{
				{PrivatelinkConnectionID: "plc-1", PSCConnectionID: "psc1", State: "pending-user-approval"},
				{PrivatelinkConnectionID: "plc-2", PSCConnectionID: "psc2", State: "pending-user-approval"},
			},
		}

		d := newApprovalResourceData(t, "10.0.0.2")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections found")
	})
}

func TestWaitForGCPConnectionState(t *testing.T) {
	useInstantStateChangeTimings(t)

	t.Run("Propagates Refresh error", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.refreshErr = fmt.Errorf("refresh failed")

		conf := waitForGCPConnectionState(
			t.Context(),
			fake,
			testProject,
			testService,
			"",
			time.Second,
			[]string{gcpPSCApprovalNotReady},
			[]string{"active"},
		)

		_, _, err := conf.Refresh()
		require.Error(t, err)
		require.Contains(t, err.Error(), "refresh failed")
	})

	t.Run("Propagates ConnectionsList error", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.connectionsListErr = fmt.Errorf("list failed")

		conf := waitForGCPConnectionState(
			t.Context(),
			fake,
			testProject,
			testService,
			"",
			time.Second,
			[]string{gcpPSCApprovalNotReady},
			[]string{"active"},
		)

		_, _, err := conf.Refresh()
		require.Error(t, err)
		require.Contains(t, err.Error(), "list failed")
	})

	t.Run("Waits when no connections exist yet", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)
		fake.emptyConnectionsListCallsRemaining = 1

		conf := waitForGCPConnectionState(
			t.Context(),
			fake,
			testProject,
			testService,
			"",
			time.Second,
			[]string{gcpPSCApprovalNotReady},
			[]string{"active"},
		)

		obj, state, err := conf.Refresh()
		require.NoError(t, err)
		require.Nil(t, obj)
		require.Equal(t, gcpPSCApprovalNotReady, state)
	})

	t.Run("Waits when PSC selector is set but no matching connection exists yet", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "active"},
		})

		conf := waitForGCPConnectionState(
			t.Context(),
			fake,
			testProject,
			testService,
			"pscX",
			time.Second,
			[]string{gcpPSCApprovalNotReady},
			[]string{"active"},
		)

		obj, state, err := conf.Refresh()
		require.NoError(t, err)
		require.Nil(t, obj)
		require.Equal(t, gcpPSCApprovalNotReady, state)
	})

	t.Run("Waits when selected PSC connection does not exist yet", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
		})

		conf := waitForGCPConnectionState(
			t.Context(),
			fake,
			testProject,
			testService,
			"psc2",
			time.Second,
			[]string{gcpPSCApprovalNotReady},
			[]string{"active"},
		)

		obj, state, err := conf.Refresh()
		require.NoError(t, err)
		require.Nil(t, obj)
		require.Equal(t, gcpPSCApprovalNotReady, state)
	})
}

func TestGCPPrivatelinkConnectionApprovalRead(t *testing.T) {
	t.Run("Populates state when privatelink_connection_id is missing and there is exactly one connection", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "active",
				UserIPAddress:           "10.0.0.2",
			},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, nil)
		d.SetId(testProject + "/" + testService)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
		require.Equal(t, testProject, d.Get("project").(string))
		require.Equal(t, testService, d.Get("service_name").(string))
		require.Equal(t, "plc1", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc1", d.Get("psc_connection_id").(string))
		require.Equal(t, "active", d.Get("state").(string))
		require.Equal(t, "10.0.0.2", d.Get("user_ip_address").(string))
	})

	t.Run("Normalizes legacy 3-part ID and uses PSC selector when multiple connections exist", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "active"},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, nil)
		d.SetId(testProject + "/" + testService + "/psc2")

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
		require.Equal(t, testProject+"/"+testService, d.Id())
		require.Equal(t, testProject, d.Get("project").(string))
		require.Equal(t, testService, d.Get("service_name").(string))
		require.Equal(t, "plc2", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc2", d.Get("psc_connection_id").(string))
	})

	t.Run("Populates state when privatelink_connection_id is missing and PSC selector matches exactly one connection", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{
				PrivatelinkConnectionID: "plc1",
				PSCConnectionID:         "psc1",
				State:                   "active",
				UserIPAddress:           "10.0.0.2",
			},
			{
				PrivatelinkConnectionID: "plc2",
				PSCConnectionID:         "psc2",
				State:                   "pending-user-approval",
			},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
			"psc_connection_id": "psc2",
		})
		d.SetId(testProject + "/" + testService)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
		require.Equal(t, "plc2", d.Get("privatelink_connection_id").(string))
		require.Equal(t, "psc2", d.Get("psc_connection_id").(string))
		require.Equal(t, "pending-user-approval", d.Get("state").(string))
	})

	t.Run("Errors when privatelink_connection_id is missing and PSC selector isn't found among multiple connections", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "active"},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
			"psc_connection_id": "pscX",
		})
		d.SetId(testProject + "/" + testService)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "psc_connection_id \"pscX\" not found")
	})

	t.Run("Errors when privatelink_connection_id is missing and PSC selector matches multiple connections", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc1", State: "active"},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
			"psc_connection_id": "psc1",
		})
		d.SetId(testProject + "/" + testService)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections match psc_connection_id")
	})

	t.Run("Errors when privatelink_connection_id is missing and multiple connections exist without PSC selector", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "active"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "active"},
		})

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, nil)
		d.SetId(testProject + "/" + testService)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "multiple privatelink connections found; set psc_connection_id to select one")
	})

	t.Run("Clears ID on 404", func(t *testing.T) {
		fake := newFakeGCPPrivatelink(nil)

		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, map[string]interface{}{
			"privatelink_connection_id": "plc1",
		})
		d.SetId(testProject + "/" + testService)
		// We need IsNewResource() == false to let ResourceReadHandleNotFound clear the ID on 404.
		// schema.ResourceData keeps this flag in an unexported field (`isNew`), so it's time for the black reflect-magic.
		field := reflect.ValueOf(d).Elem().FieldByName("isNew")
		reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().SetBool(false)

		diags := resourceGCPPrivatelinkConnectionApprovalRead(t.Context(), d, fake)
		require.False(t, diags.HasError())
		require.Empty(t, d.Id())
	})
}

func TestGCPPrivatelinkConnectionApprovalImport(t *testing.T) {
	t.Run("Parses 2-part ID", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, nil)
		d.SetId("p/s")

		_, err := resourceGCPPrivatelinkConnectionApprovalImport(t.Context(), d, nil)
		require.NoError(t, err)
		require.Equal(t, "p/s", d.Id())
		require.Equal(t, "p", d.Get("project").(string))
		require.Equal(t, "s", d.Get("service_name").(string))
		require.Empty(t, d.Get("psc_connection_id").(string))
	})

	t.Run("Parses 3-part ID and stores selector", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, aivenGCPPrivatelinkConnectionApprovalSchema, nil)
		d.SetId("p/s/psc1")

		_, err := resourceGCPPrivatelinkConnectionApprovalImport(t.Context(), d, nil)
		require.NoError(t, err)
		require.Equal(t, "p/s", d.Id())
		require.Equal(t, "p", d.Get("project").(string))
		require.Equal(t, "s", d.Get("service_name").(string))
		require.Equal(t, "psc1", d.Get("psc_connection_id").(string))
	})
}
