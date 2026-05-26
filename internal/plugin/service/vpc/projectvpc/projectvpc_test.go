package projectvpc

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestCreateViewSendsEmptyPeeringConnections(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceDataFromMaps(
		resourceSchemaInternal(),
		idFields(),
		map[string]any{
			"project":      "example-project",
			"cloud_name":   "google-europe-west2",
			"network_cidr": "192.168.0.0/24",
		},
		nil,
		nil,
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

func TestDeleteViewDeletesOnceThenWaits(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newProjectVPCResourceData(t)

	client.EXPECT().
		VpcDelete(ctx, "example-project", "example-vpc-id").
		Return(&vpc.VpcDeleteOut{State: vpc.VpcStateTypeDeleting}, nil).
		Once()
	client.EXPECT().
		VpcGet(ctx, "example-project", "example-vpc-id").
		Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeDeleted}, nil).
		Once()

	require.NoError(t, deleteView(ctx, client, d))
}

func TestWaitForDeletionReturnsNilWhenAlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newProjectVPCResourceData(t)

	client.EXPECT().
		VpcGet(ctx, "example-project", "example-vpc-id").
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Once()

	require.NoError(t, waitForDeletion(ctx, client, d))
}

func newProjectVPCResourceData(t *testing.T) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceDataFromMaps(
		resourceSchemaInternal(),
		idFields(),
		nil,
		map[string]any{
			"id":             "example-project/example-vpc-id",
			"project":        "example-project",
			"project_vpc_id": "example-vpc-id",
		},
		nil,
	)
	require.NoError(t, err)

	return d
}
