package organizationvpc

import (
	"context"
	"net/http"
	"testing"
	"time"

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

func TestDeleteViewDeletesOnceThenWaits(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newOrganizationVPCResourceData(t)

	orgID := d.Get("organization_id").(string)
	vpcID := d.Get("organization_vpc_id").(string)

	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		OrganizationVpcDelete(ctx, orgID, vpcID).
		Return(nil, avngen.Error{Status: http.StatusConflict, Message: "VPC in use"}).
		Once()

	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		OrganizationVpcDelete(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcDeleteOut{State: orgvpc.OrganizationVpcStateTypeDeleting}, nil).
		Once()

	// The organization VPC enters the DELETING state, but is not immediately deleted.
	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeDeleting}, nil).
		Once()

	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeDeleted}, nil).
		Once()

	require.NoError(t, deleteViewInternal(ctx, client, d, time.Millisecond))
}

func TestDeleteViewFailsFastOnNonRetryableDeleteError(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newOrganizationVPCResourceData(t)

	orgID := d.Get("organization_id").(string)
	vpcID := d.Get("organization_vpc_id").(string)

	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		OrganizationVpcDelete(ctx, orgID, vpcID).
		Return(nil, avngen.Error{Status: http.StatusForbidden, Message: "forbidden"}).
		Once()

	err := deleteViewInternal(ctx, client, d, time.Millisecond)
	require.Error(t, err)

	var apiErr avngen.Error
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusForbidden, apiErr.Status)
}

func TestDeleteViewReturnsNilWhenAlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newOrganizationVPCResourceData(t)

	client.EXPECT().
		OrganizationVpcGet(ctx, "example-org-id", "example-vpc-id").
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Once()

	require.NoError(t, deleteView(ctx, client, d))
}

func TestDeleteViewReturnsNilWhenDeleteFindsNoVPC(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d := newOrganizationVPCResourceData(t)

	orgID := d.Get("organization_id").(string)
	vpcID := d.Get("organization_vpc_id").(string)

	client.EXPECT().
		OrganizationVpcGet(ctx, orgID, vpcID).
		Return(&orgvpc.OrganizationVpcGetOut{State: orgvpc.OrganizationVpcStateTypeActive}, nil).
		Once()

	client.EXPECT().
		OrganizationVpcDelete(ctx, orgID, vpcID).
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Once()

	require.NoError(t, deleteViewInternal(ctx, client, d, time.Millisecond))
}

func newOrganizationVPCResourceData(t *testing.T) adapter.ResourceData {
	t.Helper()

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

	return d
}
