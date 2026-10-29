package vpc

import (
	"errors"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/require"
)

func TestSweepProjectPeerings(t *testing.T) {
	client := avngen.NewMockClient(t)
	networks := []vpc.VpcOut{
		{ProjectVpcId: "aws", CloudName: "aws-eu-west-1"},
		{ProjectVpcId: "hipaa", CloudName: "hipaa-aws-eu-west-1"},
		{ProjectVpcId: "timescale", CloudName: "timescale-aws-eu-west-1"},
		{ProjectVpcId: "aws-byoc", CloudName: "custom-example-aws-eu-west-1"},
		{ProjectVpcId: "azure", CloudName: "azure-westeurope"},
		{ProjectVpcId: "azure-byoc", CloudName: "custom-example-azure-eastus"},
		{ProjectVpcId: "gcp", CloudName: "google-europe-west1"},
		{ProjectVpcId: "upcloud", CloudName: "upcloud-fi-hel"},
	}
	client.EXPECT().VpcList(t.Context(), "project").Return(networks, nil).Once()
	for _, id := range []string{"aws", "hipaa", "timescale", "aws-byoc"} {
		client.EXPECT().VpcGet(t.Context(), "project", id).Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{
				{PeerCloudAccount: "account", PeerVpc: "peer", PeerRegion: new("eu-west-1"), VpcPeeringConnectionType: vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection},
				{PeerCloudAccount: "account", PeerVpc: "tgw", PeerRegion: new("us-east-1"), VpcPeeringConnectionType: vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment},
			},
		}, nil).Once()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", id, "account", "peer", "eu-west-1").
			Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", id, "account", "tgw", "us-east-1").
			Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()
	}
	for _, id := range []string{"azure", "azure-byoc"} {
		client.EXPECT().VpcGet(t.Context(), "project", id).Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{
				{PeerCloudAccount: "subscription", PeerVpc: "peer", PeerResourceGroup: "group-1"},
				{PeerCloudAccount: "subscription", PeerVpc: "peer", PeerResourceGroup: "group-2"},
			},
		}, nil).Once()
		for _, group := range []string{"group-1", "group-2"} {
			client.EXPECT().VpcPeeringConnectionWithResourceGroupDelete(t.Context(), "project", id, "subscription", group, "peer").
				Return(&vpc.VpcPeeringConnectionWithResourceGroupDeleteOut{}, nil).Once()
		}
	}
	for _, id := range []string{"gcp", "upcloud"} {
		client.EXPECT().VpcGet(t.Context(), "project", id).Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{{PeerCloudAccount: "account", PeerVpc: "peer"}},
		}, nil).Once()
		client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", id, "account", "peer").
			Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
	}
	require.NoError(t, sweepProjectPeerings(t.Context(), client, "project"))
}

func TestSweepProjectPeeringsMissingResources(t *testing.T) {
	client := avngen.NewMockClient(t)
	notFound := avngen.Error{Status: http.StatusNotFound, Message: "not found"}
	client.EXPECT().VpcList(t.Context(), "project").Return([]vpc.VpcOut{
		{ProjectVpcId: "missing", CloudName: "google-europe-west1"},
		{ProjectVpcId: "remaining", CloudName: "google-europe-west1"},
	}, nil).Once()
	client.EXPECT().VpcGet(t.Context(), "project", "missing").Return(nil, notFound).Once()
	client.EXPECT().VpcGet(t.Context(), "project", "remaining").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{
			{PeerCloudAccount: "account", PeerVpc: "missing"},
			{PeerCloudAccount: "account", PeerVpc: "remaining", PeerRegion: new("")},
		},
	}, nil).Once()
	client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "remaining", "account", "missing").
		Return(nil, notFound).Once()
	client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "remaining", "account", "remaining").
		Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
	require.NoError(t, sweepProjectPeerings(t.Context(), client, "project"))
}

func TestSweepProjectPeeringsSkipsNonDeletableStates(t *testing.T) {
	conflict := avngen.Error{Status: http.StatusConflict, Message: "Peering connection cannot be deleted in its current state"}

	t.Run("DELETING", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().VpcList(t.Context(), "project").Return([]vpc.VpcOut{
			{ProjectVpcId: "aws", CloudName: "aws-eu-west-1"},
		}, nil).Once()
		client.EXPECT().VpcGet(t.Context(), "project", "aws").Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{
				{PeerCloudAccount: "account", PeerVpc: "deleting", PeerRegion: new("eu-west-1"), State: "DELETING"},
				{PeerCloudAccount: "account", PeerVpc: "active", PeerRegion: new("eu-west-1"), State: "ACTIVE"},
			},
		}, nil).Once()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "aws", "account", "deleting", "eu-west-1").
			Return(nil, conflict).Maybe()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "aws", "account", "active", "eu-west-1").
			Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()

		err := sweepProjectPeerings(t.Context(), client, "project")
		client.AssertNotCalled(t, "VpcPeeringConnectionWithRegionDelete", t.Context(), "project", "aws", "account", "deleting", "eu-west-1")
		require.NoError(t, err)
	})

	t.Run("APPROVED_PEER_REQUESTED", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().VpcList(t.Context(), "project").Return([]vpc.VpcOut{
			{ProjectVpcId: "aws", CloudName: "aws-eu-west-1"},
		}, nil).Once()
		client.EXPECT().VpcGet(t.Context(), "project", "aws").Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{
				{PeerCloudAccount: "account", PeerVpc: "peer-requested", PeerRegion: new("eu-west-1"), State: "APPROVED_PEER_REQUESTED"},
				{PeerCloudAccount: "account", PeerVpc: "active", PeerRegion: new("eu-west-1"), State: "ACTIVE"},
			},
		}, nil).Once()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "aws", "account", "peer-requested", "eu-west-1").
			Return(nil, conflict).Maybe()
		client.EXPECT().VpcPeeringConnectionWithRegionDelete(t.Context(), "project", "aws", "account", "active", "eu-west-1").
			Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).Once()

		err := sweepProjectPeerings(t.Context(), client, "project")
		client.AssertNotCalled(t, "VpcPeeringConnectionWithRegionDelete", t.Context(), "project", "aws", "account", "peer-requested", "eu-west-1")
		require.NoError(t, err)
	})
}

func TestSweepProjectPeeringsContinuesAfterErrors(t *testing.T) {
	client := avngen.NewMockClient(t)
	readError := errors.New("VPC lookup failed")
	deleteError := errors.New("peering deletion failed")
	client.EXPECT().VpcList(t.Context(), "project").Return([]vpc.VpcOut{
		{ProjectVpcId: "unreadable", CloudName: "aws-eu-west-1"},
		{ProjectVpcId: "azure", CloudName: "azure-westeurope"},
		{ProjectVpcId: "gcp", CloudName: "google-europe-west1"},
	}, nil).Once()
	client.EXPECT().VpcGet(t.Context(), "project", "unreadable").Return(nil, readError).Once()
	client.EXPECT().VpcGet(t.Context(), "project", "azure").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{
			{PeerCloudAccount: "subscription", PeerVpc: "peer", PeerResourceGroup: "group"},
			{PeerCloudAccount: "subscription", PeerVpc: "remaining", PeerResourceGroup: "group"},
		},
	}, nil).Once()
	client.EXPECT().VpcPeeringConnectionWithResourceGroupDelete(t.Context(), "project", "azure", "subscription", "group", "peer").
		Return(nil, deleteError).Once()
	client.EXPECT().VpcPeeringConnectionWithResourceGroupDelete(t.Context(), "project", "azure", "subscription", "group", "remaining").
		Return(&vpc.VpcPeeringConnectionWithResourceGroupDeleteOut{}, nil).Once()
	client.EXPECT().VpcGet(t.Context(), "project", "gcp").Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{{PeerCloudAccount: "account", PeerVpc: "peer"}},
	}, nil).Once()
	client.EXPECT().VpcPeeringConnectionDelete(t.Context(), "project", "gcp", "account", "peer").
		Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).Once()
	err := sweepProjectPeerings(t.Context(), client, "project")
	require.ErrorIs(t, err, readError)
	require.ErrorIs(t, err, deleteError)
}
