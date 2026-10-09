package projectvpc

import (
	"context"
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
