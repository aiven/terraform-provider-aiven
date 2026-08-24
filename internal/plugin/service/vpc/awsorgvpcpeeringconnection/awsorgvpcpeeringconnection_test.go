package awsorgvpcpeeringconnection

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testOrganizationID          = "example-organization-id"
	testOrganizationVPCID       = "example-organization-vpc-id"
	testAWSAccountID            = "123456789012"
	testAWSVPCID                = "vpc-1234567890abcdef0"
	testAWSRegion               = "eu-central-1"
	testConnectionID            = "example-connection-id"
	testStaleConnectionID       = "stale-connection-id"
	testAWSVPCPeeringConnection = "pcx-1234567890abcdef0"
)

func TestFlattenModifier(t *testing.T) {
	t.Run("flattens AWS VPC peering connection ID", func(t *testing.T) {
		d := newTestResourceData(t)
		dto := map[string]any{
			"state": string(organizationvpc.VpcPeeringConnectionStateTypeActive),
			"state_info": map[string]any{
				"aws_vpc_peering_connection_id": testAWSVPCPeeringConnection,
			},
		}

		require.NoError(t, flattenModifier(t.Context(), nil)(d, dto))
		require.Equal(t, testAWSVPCPeeringConnection, dto["aws_vpc_peering_connection_id"])
	})

	t.Run("clears AWS VPC peering connection ID when state info omits it", func(t *testing.T) {
		d := newTestResourceData(t)
		require.NoError(t, d.Set("aws_vpc_peering_connection_id", testAWSVPCPeeringConnection))
		dto := map[string]any{
			"state":      string(organizationvpc.VpcPeeringConnectionStateTypeActive),
			"state_info": map[string]any{},
		}

		require.NoError(t, flattenModifier(t.Context(), nil)(d, dto))
		_, ok := d.GetOk("aws_vpc_peering_connection_id")
		require.False(t, ok)
	})
}

func TestDeleteView(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newTestResourceData(t)
	require.NoError(t, d.Set("peering_connection_id", testStaleConnectionID))

	client.EXPECT().
		OrganizationVpcGet(t.Context(), testOrganizationID, testOrganizationVPCID).
		Return(&organizationvpc.OrganizationVpcGetOut{
			PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
				{
					PeerCloudAccount:    "another-account-id",
					PeerVpc:             testAWSVPCID,
					PeerRegion:          new(testAWSRegion),
					PeeringConnectionId: new("another-account-connection-id"),
				},
				{
					PeerCloudAccount:    testAWSAccountID,
					PeerVpc:             "another-vpc-id",
					PeerRegion:          new(testAWSRegion),
					PeeringConnectionId: new("another-vpc-connection-id"),
				},
				{
					PeerCloudAccount:    testAWSAccountID,
					PeerVpc:             testAWSVPCID,
					PeerRegion:          new("another-region"),
					PeeringConnectionId: new("another-region-connection-id"),
				},
				{
					PeerCloudAccount:    testAWSAccountID,
					PeerVpc:             testAWSVPCID,
					PeeringConnectionId: new("missing-region-connection-id"),
				},
				{
					PeerCloudAccount:    testAWSAccountID,
					PeerVpc:             testAWSVPCID,
					PeerRegion:          new(testAWSRegion),
					PeeringConnectionId: new(testConnectionID),
				},
			},
		}, nil).
		Once()
	client.EXPECT().
		OrganizationVpcPeeringConnectionDeleteById(t.Context(), testOrganizationID, testOrganizationVPCID, testConnectionID).
		Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).
		Once()

	require.NoError(t, deleteView(t.Context(), client, d))
}

func newTestResourceData(t *testing.T) adapter.ResourceData {
	t.Helper()

	values := map[string]any{
		"organization_id":       testOrganizationID,
		"organization_vpc_id":   testOrganizationVPCID,
		"aws_account_id":        testAWSAccountID,
		"aws_vpc_id":            testAWSVPCID,
		"aws_vpc_region":        testAWSRegion,
		"id":                    testOrganizationID + "/" + testOrganizationVPCID + "/" + testAWSAccountID + "/" + testAWSVPCID + "/" + testAWSRegion,
		"peering_connection_id": testConnectionID,
	}

	d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(values))
	require.NoError(t, err)
	return d
}
