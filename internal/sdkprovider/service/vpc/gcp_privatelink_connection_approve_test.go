package vpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
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

type fakeGCPPrivatelink struct {
	project string
	service string

	connections   []aiven.GCPPrivatelinkConnectionResponse
	approveCalls  []string
	approveUserIP []string

	connectionsListCalls                int
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
	return f.assertProjectService(project, serviceName)
}

func (f *fakeGCPPrivatelink) ConnectionsList(
	_ context.Context,
	project,
	serviceName string,
) (*aiven.GCPPrivatelinkConnectionsResponse, error) {
	if err := f.assertProjectService(project, serviceName); err != nil {
		return nil, err
	}

	f.connectionsListCalls++
	if f.emptyConnectionsListCallsRemaining > 0 {
		f.emptyConnectionsListCallsRemaining--
		f.returnedEmptyConnectionsListAtLeast = true
		return &aiven.GCPPrivatelinkConnectionsResponse{Connections: nil}, nil
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

	t.Run("Errors when multiple connections exist", func(t *testing.T) {
		fake := newFakeGCPPrivatelink([]aiven.GCPPrivatelinkConnectionResponse{
			{PrivatelinkConnectionID: "plc1", PSCConnectionID: "psc1", State: "connected", UserIPAddress: "10.0.0.2"},
			{PrivatelinkConnectionID: "plc2", PSCConnectionID: "psc2", State: "pending-user-approval"},
		})

		d := newApprovalResourceData(t, "10.0.0.3")

		diags := resourceGCPPrivatelinkConnectionApprovalUpdate(t.Context(), d, fake)
		require.True(t, diags.HasError(), "expected error, got: %#v", diags)
		require.Contains(t, diags[0].Summary, "number of privatelink connections != 1")
	})

	t.Run("Errors when another connection exists", func(t *testing.T) {
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
		require.Contains(t, diags2[0].Summary, "number of privatelink connections != 1")
	})
}

func TestGCPPrivatelinkConnectionApprovalRead(t *testing.T) {
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
