package projectvpc

import (
	"context"
	"net/http"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestCreateViewSendsEmptyPeeringConnections(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(map[string]any{
			"project":      "example-project",
			"cloud_name":   "google-europe-west2",
			"network_cidr": "192.168.0.0/24",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		VpcCreate(ctx, "example-project", mock.MatchedBy(func(req *vpc.VpcCreateIn) bool {
			return req.CloudName == "google-europe-west2" &&
				req.NetworkCidr == "192.168.0.0/24" &&
				req.PeeringConnections != nil &&
				len(req.PeeringConnections) == 0
		})).
		Return(&vpc.VpcCreateOut{
			CloudName:          "google-europe-west2",
			NetworkCidr:        "192.168.0.0/24",
			ProjectVpcId:       "example-vpc-id",
			State:              vpc.VpcStateTypeActive,
			PeeringConnections: []vpc.PeeringConnectionOut{},
		}, nil).
		Once()

	require.NoError(t, createView(ctx, client, d))
}

// TestDeleteViewDeletesOnceThenWaits starts with the VPC in the ACTIVE state.
// It first attempts to delete the VPC, but receives a conflict error indicating the VPC is still in use.
// It tries to delete the VPC again and the operation succeeds after a short delay.
// The VPC enters the DELETING state.
// Finally, the VPC transitions to the DELETED state.
func TestDeleteViewDeletesOnceThenWaits(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newProjectVPCResourceData(t)

	projectName := d.Get("project").(string)
	vpcID := d.Get("project_vpc_id").(string)

	client.EXPECT().
		VpcGet(ctx, projectName, vpcID).
		Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		VpcDelete(ctx, projectName, vpcID).
		Return(nil, avngen.Error{Status: http.StatusConflict, Message: "VPC in use"}).
		Once()

	client.EXPECT().
		VpcGet(ctx, projectName, vpcID).
		Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		VpcDelete(ctx, projectName, vpcID).
		Return(&vpc.VpcDeleteOut{State: vpc.VpcStateTypeDeleting}, nil).
		Once()

	// The VPC enters the DELETING state, but not immediately deleted.
	client.EXPECT().
		VpcGet(ctx, projectName, vpcID).
		Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeDeleting}, nil).
		Once()

	client.EXPECT().
		VpcGet(ctx, projectName, vpcID).
		Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeDeleted}, nil).
		Once()

	require.NoError(t, deleteViewInternal(ctx, client, d, 1*time.Millisecond))
}

func TestWaitForDeletionReturnsNilWhenAlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newProjectVPCResourceData(t)

	client.EXPECT().
		VpcGet(ctx, "example-project", "example-vpc-id").
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Once()

	require.NoError(t, deleteView(ctx, client, d))
}

func newProjectVPCResourceData(t *testing.T) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestState(map[string]any{
			"id":             "example-project/example-vpc-id",
			"project":        "example-project",
			"project_vpc_id": "example-vpc-id",
		}),
	)
	require.NoError(t, err)

	return d
}
