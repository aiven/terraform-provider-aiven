package vpc

import (
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/stretchr/testify/require"
)

func TestProjectPeeringCreateRejectsExisting(t *testing.T) {
	check := func(t *testing.T, cloud string, request *vpc.VpcPeeringConnectionCreateIn, connection vpc.PeeringConnectionOut) {
		t.Helper()
		client := avngen.NewMockClient(t)
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").
			Return(&vpc.VpcGetOut{CloudName: cloud, PeeringConnections: []vpc.PeeringConnectionOut{connection}}, nil).Once()
		id := &ProjectPeeringID{Project: "project", ProjectVPCID: "vpc", PeerCloudAccount: request.PeerCloudAccount, PeerVPC: request.PeerVpc, PeerRegion: request.PeerRegion}
		err := id.Create(t.Context(), client, request)
		require.ErrorContains(t, err, "already exists in state "+string(connection.State))
		require.ErrorContains(t, err, "destroy-before-create")
		client.AssertNumberOfCalls(t, "VpcPeeringConnectionCreate", 0)
	}

	t.Run("pending GCP connection", func(t *testing.T) {
		check(t, "google-europe-west10", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "network", State: vpc.VpcPeeringConnectionStateTypePendingPeer})
	})

	t.Run("failed GCP connection", func(t *testing.T) {
		check(t, "google-europe-west10", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "network", State: vpc.VpcPeeringConnectionStateTypeRejectedByPeer})
	})

	t.Run("deleting GCP connection", func(t *testing.T) {
		check(t, "google-europe-west10", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "network", State: vpc.VpcPeeringConnectionStateTypeDeleting})
	})

	t.Run("AWS default region", func(t *testing.T) {
		check(t, "aws-eu-west-1", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "vpc-peer"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("eu-west-1"), State: vpc.VpcPeeringConnectionStateTypeActive})
	})

	t.Run("AWS explicit region", func(t *testing.T) {
		check(t, "aws-eu-west-1", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("us-east-1"),
		}, vpc.PeeringConnectionOut{
			PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("us-east-1"), State: vpc.VpcPeeringConnectionStateTypeActive,
		})
	})

	t.Run("Azure credentials do not change the key", func(t *testing.T) {
		check(t, "azure-westeurope", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: new("group"),
			PeerAzureAppId: new("new-app"), PeerAzureTenantId: new("new-tenant"),
		}, vpc.PeeringConnectionOut{
			PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: "group",
			PeerAzureAppId: "old-app", PeerAzureTenantId: "old-tenant", State: vpc.VpcPeeringConnectionStateTypeActive,
		})
	})

	t.Run("TGW CIDRs do not change the key", func(t *testing.T) {
		check(t, "aws-eu-west-1", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "account", PeerVpc: "tgw-peer", UserPeerNetworkCidrs: &[]string{"10.1.0.0/24"},
		}, vpc.PeeringConnectionOut{
			PeerCloudAccount: "account", PeerVpc: "tgw-peer", PeerRegion: new("eu-west-1"),
			UserPeerNetworkCidrs: []string{"10.0.0.0/24"}, State: vpc.VpcPeeringConnectionStateTypeActive,
		})
	})
}

func TestProjectPeeringCreateAllowsDistinctConnections(t *testing.T) {
	check := func(t *testing.T, cloud string, request *vpc.VpcPeeringConnectionCreateIn, connections ...vpc.PeeringConnectionOut) {
		t.Helper()
		client := avngen.NewMockClient(t)
		lookup := client.EXPECT().VpcGet(t.Context(), "project", "vpc").
			Return(&vpc.VpcGetOut{CloudName: cloud, PeeringConnections: connections}, nil).Once()
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "project", "vpc", request).
			Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).Once().NotBefore(lookup)
		id := &ProjectPeeringID{Project: "project", ProjectVPCID: "vpc", PeerCloudAccount: request.PeerCloudAccount, PeerVPC: request.PeerVpc, PeerRegion: request.PeerRegion}
		require.NoError(t, id.Create(t.Context(), client, request))
	}

	t.Run("empty VPC", func(t *testing.T) {
		check(t, "google-europe-west10", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"})
	})

	t.Run("different accounts and networks", func(t *testing.T) {
		check(t, "google-europe-west10", &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "other-account", PeerVpc: "network"},
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "other-network"})
	})

	t.Run("AWS omitted region is not a wildcard", func(t *testing.T) {
		request := &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "vpc-peer"}
		check(t, "aws-eu-west-1", request,
			vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("us-east-1")})
		require.Nil(t, request.PeerRegion, "checking the default region must preserve omission in the POST")
	})

	t.Run("AWS explicit region overrides the VPC region", func(t *testing.T) {
		check(t, "aws-eu-west-1", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("us-east-1"),
		}, vpc.PeeringConnectionOut{PeerCloudAccount: "account", PeerVpc: "vpc-peer", PeerRegion: new("eu-west-1")})
	})

	t.Run("Azure different resource groups", func(t *testing.T) {
		check(t, "azure-westeurope", &vpc.VpcPeeringConnectionCreateIn{
			PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: new("new-group"),
		}, vpc.PeeringConnectionOut{PeerCloudAccount: "subscription", PeerVpc: "network", PeerResourceGroup: "old-group"})
	})
}

func TestProjectPeeringCreateErrors(t *testing.T) {
	id := &ProjectPeeringID{Project: "project", ProjectVPCID: "vpc", PeerCloudAccount: "account", PeerVPC: "network"}
	request := &vpc.VpcPeeringConnectionCreateIn{PeerCloudAccount: "account", PeerVpc: "network"}

	t.Run("failed lookup does not create", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		lookupError := &avngen.Error{Status: http.StatusForbidden, Message: "forbidden"}
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(nil, lookupError).Once()
		require.ErrorIs(t, id.Create(t.Context(), client, request), lookupError)
		client.AssertNumberOfCalls(t, "VpcPeeringConnectionCreate", 0)
	})

	t.Run("missing parent VPC does not create", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		lookupError := &avngen.Error{Status: http.StatusNotFound, Message: "VPC not found"}
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(nil, lookupError).Once()
		require.ErrorIs(t, id.Create(t.Context(), client, request), lookupError)
		client.AssertNumberOfCalls(t, "VpcPeeringConnectionCreate", 0)
	})

	t.Run("POST conflict remains an error", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		conflict := &avngen.Error{Status: http.StatusConflict, Message: "VPC peering connection already exists"}
		client.EXPECT().VpcGet(t.Context(), "project", "vpc").Return(&vpc.VpcGetOut{}, nil).Once()
		client.EXPECT().VpcPeeringConnectionCreate(t.Context(), "project", "vpc", request).Return(nil, conflict).Once()
		require.ErrorIs(t, id.Create(t.Context(), client, request), conflict)
	})
}
