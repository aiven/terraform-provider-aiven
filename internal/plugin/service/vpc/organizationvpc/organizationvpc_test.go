package organizationvpc

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	orgvpc "github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestCreateViewSendsCloudsAndEmptyPeeringConnections(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(map[string]any{
			"organization_id": "example-org-id",
			"cloud_name":      "aws-eu-west-1",
			"network_cidr":    "10.0.0.0/24",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		OrganizationVpcCreate(ctx, "example-org-id", mock.MatchedBy(func(req *orgvpc.OrganizationVpcCreateIn) bool {
			return len(req.Clouds) == 1 &&
				req.Clouds[0].CloudName == "aws-eu-west-1" &&
				req.Clouds[0].NetworkCidr == "10.0.0.0/24" &&
				req.PeeringConnections != nil &&
				len(req.PeeringConnections) == 0
		})).
		Return(&orgvpc.OrganizationVpcCreateOut{
			Clouds: []orgvpc.CloudOut{{
				CloudName:   "aws-eu-west-1",
				NetworkCidr: "10.0.0.0/24",
			}},
			OrganizationId:    "example-org-id",
			OrganizationVpcId: "example-vpc-id",
			State:             orgvpc.OrganizationVpcStateTypeActive,
		}, nil).
		Once()

	require.NoError(t, createView(ctx, client, d))
}

func TestFlattenModifierRequiresOneCloud(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestState(map[string]any{
			"id":                  "example-org-id/example-vpc-id",
			"organization_id":     "example-org-id",
			"organization_vpc_id": "example-vpc-id",
		}),
	)
	require.NoError(t, err)

	err = d.Flatten(&orgvpc.OrganizationVpcGetOut{
		Clouds:            []orgvpc.CloudOut{},
		OrganizationId:    "example-org-id",
		OrganizationVpcId: "example-vpc-id",
		State:             orgvpc.OrganizationVpcStateTypeActive,
	}, flattenModifier(ctx, client))
	require.EqualError(t, err, "expected exactly 1 cloud, got 0")
}
