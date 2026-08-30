package transitgatewayvpcattachment_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

const (
	gcpVPCPeeringTestEnv = "GCP_VPC_PEERING_TEST"
	defaultGCPRegion     = "europe-west10"
)

type gcpConfig struct {
	AivenProject string
	GCPProject   string
	Region       string
}

func TestAccAivenGCPTransitGatewayVPCAttachment_basic(t *testing.T) {
	// Despite its legacy AWS-specific name, the resource uses the generic project
	// VPC peering API. This scenario exercises its regionless GCP path.
	config := getGCPConfig(t)
	prefix := "test-tf-acc-" + acctest.RandString(7)
	networkName := prefix + "-network"
	fullConfig := testAccAivenGCPTransitGatewayVPCAttachment(prefix, networkName, config)
	ctx := t.Context()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": gcpExternalProvider(),
		},
		Steps: []resource.TestStep{
			{
				Config: fullConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccAivenGCPTransitGatewayVPCAttachmentChecks(config),
					resource.TestCheckResourceAttr("google_compute_network_peering.peer", "state", "ACTIVE"),
					testAccCheckAivenGCPTransitGatewayVPCAttachmentActive(ctx),
				),
			},
			{
				// The previous check waits until Aiven observes the reciprocal Google
				// peering. Refresh once so Terraform records the resulting ACTIVE state.
				Config: fullConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccAivenGCPTransitGatewayVPCAttachmentChecks(config),
					resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "state", string(vpc.VpcPeeringConnectionStateTypeActive)),
					resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "state", string(vpc.VpcPeeringConnectionStateTypeActive)),
				),
			},
			{
				ResourceName:      transitGatewayAttachmentResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Delete both sides of the peering while the Aiven and Google VPCs
				// still exist, so deleting either network cannot hide a broken path.
				Config: testAccAivenGCPTransitGatewayVPCAttachmentFixture(networkName, config),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionDestroy),
					},
				},
				Check: testAccCheckAivenGCPTransitGatewayVPCAttachmentDeleted(ctx, config.GCPProject, networkName),
			},
		},
	})
}

func testAccAivenGCPTransitGatewayVPCAttachmentChecks(c *gcpConfig) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "vpc_id", "aiven_project_vpc.aiven_vpc", "id"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_cloud_account", c.GCPProject),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "peer_vpc", "google_compute_network.peer", "name"),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentResourceName, "peer_region"),
		resource.TestMatchResourceAttr(transitGatewayAttachmentResourceName, "state", liveAttachmentStatePattern),
		resource.TestCheckResourceAttrSet(transitGatewayAttachmentResourceName, "state_info.to_project_id"),
		resource.TestCheckResourceAttrSet(transitGatewayAttachmentResourceName, "state_info.to_vpc_network"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "user_peer_network_cidrs.#", "0"),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentResourceName, "peering_connection_id"),
		testAccCheckAivenTransitGatewayVPCAttachmentIDParts(4),

		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "id", transitGatewayAttachmentResourceName, "id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "vpc_id", transitGatewayAttachmentResourceName, "vpc_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_cloud_account", transitGatewayAttachmentResourceName, "peer_cloud_account"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_vpc", transitGatewayAttachmentResourceName, "peer_vpc"),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentDataSourceName, "peer_region"),
		resource.TestMatchResourceAttr(transitGatewayAttachmentDataSourceName, "state", liveAttachmentStatePattern),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "state_info.to_project_id", transitGatewayAttachmentResourceName, "state_info.to_project_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "state_info.to_vpc_network", transitGatewayAttachmentResourceName, "state_info.to_vpc_network"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "user_peer_network_cidrs.#", "0"),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentDataSourceName, "peering_connection_id"),
	)
}

func testAccCheckAivenGCPTransitGatewayVPCAttachmentActive(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attachment, ok := s.RootModule().Resources[transitGatewayAttachmentResourceName]
		if !ok {
			return fmt.Errorf("transit gateway VPC attachment not found in state")
		}

		parts := strings.Split(attachment.Primary.ID, "/")
		if len(parts) != 4 {
			return fmt.Errorf("parse GCP VPC peering ID: expected 4 parts, got %d", len(parts))
		}
		project, projectVPCID, peerCloudAccount, peerVPC := parts[0], parts[1], parts[2], parts[3]

		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			return fmt.Errorf("initialize Aiven client: %w", err)
		}
		wait := &retry.StateChangeConf{
			Pending: []string{
				string(vpc.VpcPeeringConnectionStateTypeApproved),
				string(vpc.VpcPeeringConnectionStateTypeApprovedPeerRequested),
				string(vpc.VpcPeeringConnectionStateTypePendingPeer),
			},
			Target: []string{string(vpc.VpcPeeringConnectionStateTypeActive)},
			Refresh: func() (any, string, error) {
				vpcOut, err := client.VpcGet(ctx, project, projectVPCID)
				if err != nil {
					return nil, "", fmt.Errorf("get Aiven project VPC: %w", err)
				}

				for _, connection := range vpcOut.PeeringConnections {
					if connection.PeerCloudAccount != peerCloudAccount || connection.PeerVpc != peerVPC {
						continue
					}
					if connection.PeerRegion != nil && *connection.PeerRegion != "" {
						return nil, "", fmt.Errorf("expected GCP VPC peering without peer region, got %q", *connection.PeerRegion)
					}
					if connection.VpcPeeringConnectionType != vpc.VpcPeeringConnectionTypeGoogleVpcPeering {
						return nil, "", fmt.Errorf("expected %q peering connection type, got %q", vpc.VpcPeeringConnectionTypeGoogleVpcPeering, connection.VpcPeeringConnectionType)
					}

					return &connection, string(connection.State), nil
				}

				return nil, "", fmt.Errorf("GCP VPC peering connection %q not found in Aiven VPC", attachment.Primary.ID)
			},
			Timeout:      10 * time.Minute,
			PollInterval: 5 * time.Second,
		}
		if _, err := wait.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("wait for active GCP VPC peering connection: %w", err)
		}

		return nil
	}
}

func testAccCheckAivenGCPTransitGatewayVPCAttachmentDeleted(ctx context.Context, peerCloudAccount, peerVPC string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		projectVPC, ok := s.RootModule().Resources["aiven_project_vpc.aiven_vpc"]
		if !ok {
			return fmt.Errorf("Aiven project VPC not found in state")
		}

		parts := strings.Split(projectVPC.Primary.ID, "/")
		if len(parts) != 2 {
			return fmt.Errorf("parse Aiven project VPC ID: expected 2 parts, got %d", len(parts))
		}

		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			return fmt.Errorf("initialize Aiven client: %w", err)
		}
		vpcOut, err := client.VpcGet(ctx, parts[0], parts[1])
		if err != nil {
			return fmt.Errorf("get Aiven project VPC: %w", err)
		}

		for _, connection := range vpcOut.PeeringConnections {
			if connection.PeerCloudAccount == peerCloudAccount && connection.PeerVpc == peerVPC {
				return fmt.Errorf("GCP VPC peering connection for network %q still exists in state %q", peerVPC, connection.State)
			}
		}

		return nil
	}
}

func testAccAivenGCPTransitGatewayVPCAttachment(prefix, networkName string, c *gcpConfig) string {
	return testAccAivenGCPTransitGatewayVPCAttachmentFixture(networkName, c) + fmt.Sprintf(`
resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.aiven_vpc.id
  peer_cloud_account = %[1]q
  peer_vpc           = google_compute_network.peer.name

  timeouts {
    create = "10m"
  }
}

resource "google_compute_network_peering" "peer" {
  name    = %[2]q
  network = google_compute_network.peer.id
  peer_network = format(
    "https://www.googleapis.com/compute/v1/projects/%%s/global/networks/%%s",
    aiven_transit_gateway_vpc_attachment.attachment.state_info["to_project_id"],
    aiven_transit_gateway_vpc_attachment.attachment.state_info["to_vpc_network"],
  )
}

data "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_transit_gateway_vpc_attachment.attachment.vpc_id
  peer_cloud_account = aiven_transit_gateway_vpc_attachment.attachment.peer_cloud_account
  peer_vpc           = aiven_transit_gateway_vpc_attachment.attachment.peer_vpc

  depends_on = [google_compute_network_peering.peer]
}
`, c.GCPProject, prefix+"-peering")
}

func testAccAivenGCPTransitGatewayVPCAttachmentFixture(networkName string, c *gcpConfig) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[1]q
}

provider "google" {
  project = %[2]q
  region  = %[3]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "google-%[3]s"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "google_compute_network" "peer" {
  name                    = %[4]q
  auto_create_subnetworks = false
}
`, c.AivenProject, c.GCPProject, c.Region, networkName)
}

func getGCPConfig(t *testing.T) *gcpConfig {
	t.Helper()

	env := acc.RequireEnvVars(t, gcpVPCPeeringTestEnv, "GOOGLE_PROJECT")
	return &gcpConfig{
		AivenProject: acc.ProjectName(),
		GCPProject:   env["GOOGLE_PROJECT"],
		Region:       defaultGCPRegion,
	}
}

func gcpExternalProvider() resource.ExternalProvider {
	return resource.ExternalProvider{
		Source:            "hashicorp/google",
		VersionConstraint: "=6.15.0",
	}
}
