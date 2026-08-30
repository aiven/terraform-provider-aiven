package transitgatewayvpcattachment_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
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
	awsTransitGatewayTestEnv = "AWS_TRANSIT_GATEWAY_TEST"
	aivenAWSAccountID        = "675999398324"
	defaultAWSRegion         = "eu-west-3"
	testPeerNetworkCIDR      = "10.0.0.0/24"
)

type awsConfig struct {
	Project string
	Region  string
}

func TestAccAivenAWSTransitGatewayVPCAttachment_basic(t *testing.T) {
	config := getAWSConfig(t)
	prefix := "test-tf-acc-" + acctest.RandString(7)
	peerRegion := config.Region
	cidrs := []string{testPeerNetworkCIDR}
	emptyCIDRs := []string{}
	ctx := t.Context()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": awsExternalProvider(),
		},
		Steps: []resource.TestStep{
			{
				// A null Optional+Computed value is Terraform's unconfigured case:
				// fresh Plugin Framework creates must not inherit the SDKv2 required-set behavior.
				Config: testAccAivenAWSTransitGatewayVPCAttachment(prefix, config, nil, &peerRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccAivenAWSTransitGatewayVPCAttachmentChecks(config, nil, 5),
					resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", config.Region),
					testAccCheckAivenAWSTransitGatewayVPCAttachmentActive(ctx, config.Region),
				),
			},
			{
				Config: testAccAivenAWSTransitGatewayVPCAttachment(prefix, config, &cidrs, &peerRegion),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccAivenAWSTransitGatewayVPCAttachmentChecks(config, cidrs, 5),
					resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", config.Region),
				),
			},
			{
				ResourceName:      transitGatewayAttachmentResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Omitting an Optional+Computed set relinquishes management; it must
				// not silently clear the CIDRs learned during the previous step.
				Config: testAccAivenAWSTransitGatewayVPCAttachment(prefix, config, nil, &peerRegion),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccAivenAWSTransitGatewayVPCAttachmentChecks(config, cidrs, 5),
					resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", config.Region),
				),
			},
			{
				Config: testAccAivenAWSTransitGatewayVPCAttachment(prefix, config, &emptyCIDRs, &peerRegion),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccAivenAWSTransitGatewayVPCAttachmentChecks(config, emptyCIDRs, 5),
					resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", config.Region),
				),
			},
			{
				// Delete the attachment while its Aiven VPC and AWS Transit Gateway
				// still exist, so their teardown cannot mask a broken Delete path.
				Config: testAccAivenAWSTransitGatewayVPCAttachmentFixture(prefix, config),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionDestroy),
					},
				},
				Check: testAccCheckAivenAWSTransitGatewayVPCAttachmentDeleted(ctx, "aws_ec2_transit_gateway.transit_gateway"),
			},
		},
	})
}

func TestAccAivenAWSTransitGatewayVPCAttachment_backwardCompat(t *testing.T) {
	config := getAWSConfig(t)
	prefix := "test-tf-acc-" + acctest.RandString(7)
	cidrs := []string{testPeerNetworkCIDR}
	ctx := t.Context()
	tfConfig := testAccAivenAWSTransitGatewayVPCAttachment(prefix, config, &cidrs, nil)
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig:           tfConfig,
		OldProviderVersion: "4.61.0",
		Checks:             testAccAivenAWSTransitGatewayVPCAttachmentChecks(config, cidrs, 4),
	})
	// The SDKv2 provider creates a four-part ID when peer_region is omitted, but
	// its Read writes the API-discovered region into the Optional-only ForceNew
	// attribute. Its post-apply plan therefore contains a known replacement.
	// The Plugin Framework provider must absorb that legacy state without a diff.
	steps[0].ExpectNonEmptyPlan = true
	steps[0].Check = resource.ComposeTestCheckFunc(
		steps[0].Check,
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", config.Region),
		testAccCheckAivenAWSTransitGatewayVPCAttachmentActive(ctx, config.Region),
	)
	steps[1].ConfigPlanChecks.PreApply = append(
		steps[1].ConfigPlanChecks.PreApply,
		plancheck.ExpectEmptyPlan(),
	)
	steps[1].Check = resource.ComposeTestCheckFunc(
		steps[1].Check,
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentResourceName, "peer_region"),
	)

	// The shared compatibility helper configures the old and current Aiven
	// providers. This fixture also creates and shares an AWS Transit Gateway, so
	// make the external AWS provider available in both steps.
	for i := range steps {
		if steps[i].ExternalProviders == nil {
			steps[i].ExternalProviders = make(map[string]resource.ExternalProvider)
		}
		steps[i].ExternalProviders["aws"] = awsExternalProvider()
	}
	steps = append(steps,
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			ExternalProviders: map[string]resource.ExternalProvider{
				"aws": awsExternalProvider(),
			},
			ResourceName:      transitGatewayAttachmentResourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			ExternalProviders: map[string]resource.ExternalProvider{
				"aws": awsExternalProvider(),
			},
			Config: testAccAivenAWSTransitGatewayVPCAttachmentFixture(prefix, config),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionDestroy),
				},
			},
			Check: testAccCheckAivenAWSTransitGatewayVPCAttachmentDeleted(ctx, "aws_ec2_transit_gateway.transit_gateway"),
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps:    steps,
	})
}

func TestAccAivenAWSTransitGatewayVPCAttachment_legacyVPCPeering(t *testing.T) {
	// This characterizes a legacy contract rather than the clean-slate design:
	// the resource still accepts an ordinary AWS vpc-* and exposes its pcx-* ID.
	config := getAWSConfig(t)
	prefix := "test-tf-acc-" + acctest.RandString(7)
	ctx := t.Context()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": awsExternalProvider(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeering(prefix, config),
				Check:  testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeeringChecks(config),
			},
			{
				ResourceName:      transitGatewayAttachmentResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeeringFixture(prefix, config),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(transitGatewayAttachmentResourceName, plancheck.ResourceActionDestroy),
					},
				},
				Check: testAccCheckAivenAWSTransitGatewayVPCAttachmentDeleted(ctx, "aws_vpc.peer"),
			},
		},
	})
}

func testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeeringChecks(c *awsConfig) resource.TestCheckFunc {
	peeringIDPattern := regexp.MustCompile(`^pcx-[[:xdigit:]]+$`)

	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "vpc_id", "aiven_project_vpc.aiven_vpc", "id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "peer_cloud_account", "aws_vpc.peer", "owner_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "peer_vpc", "aws_vpc.peer", "id"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "peer_region", c.Region),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "state", string(vpc.VpcPeeringConnectionStateTypePendingPeer)),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "user_peer_network_cidrs.#", "0"),
		resource.TestMatchResourceAttr(transitGatewayAttachmentResourceName, "peering_connection_id", peeringIDPattern),
		resource.TestCheckResourceAttrPair(
			transitGatewayAttachmentResourceName,
			"peering_connection_id",
			transitGatewayAttachmentResourceName,
			"state_info.aws_vpc_peering_connection_id",
		),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentResourceName, "state_info.aws_transit_gateway_attachment_id"),
		testAccCheckAivenTransitGatewayVPCAttachmentIDParts(5),

		// Data source
		testAccCheckAivenAWSTransitGatewayVPCAttachmentDataSourceID(c.Region),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "vpc_id", transitGatewayAttachmentResourceName, "vpc_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_cloud_account", transitGatewayAttachmentResourceName, "peer_cloud_account"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_vpc", transitGatewayAttachmentResourceName, "peer_vpc"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "peer_region", c.Region),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "state", string(vpc.VpcPeeringConnectionStateTypePendingPeer)),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "user_peer_network_cidrs.#", "0"),
		resource.TestMatchResourceAttr(transitGatewayAttachmentDataSourceName, "peering_connection_id", peeringIDPattern),
		resource.TestCheckResourceAttrPair(
			transitGatewayAttachmentDataSourceName,
			"peering_connection_id",
			transitGatewayAttachmentDataSourceName,
			"state_info.aws_vpc_peering_connection_id",
		),
		resource.TestCheckResourceAttrPair(
			transitGatewayAttachmentDataSourceName,
			"peering_connection_id",
			transitGatewayAttachmentResourceName,
			"peering_connection_id",
		),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentDataSourceName, "state_info.aws_transit_gateway_attachment_id"),
	)
}

func testAccAivenAWSTransitGatewayVPCAttachmentChecks(c *awsConfig, cidrs []string, idParts int) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "vpc_id", "aiven_project_vpc.aiven_vpc", "id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "peer_cloud_account", "aws_ec2_transit_gateway.transit_gateway", "owner_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentResourceName, "peer_vpc", "aws_ec2_transit_gateway.transit_gateway", "id"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentResourceName, "user_peer_network_cidrs.#", strconv.Itoa(len(cidrs))),
		resource.TestMatchResourceAttr(transitGatewayAttachmentResourceName, "state", liveAttachmentStatePattern),
		resource.TestCheckResourceAttrSet(transitGatewayAttachmentResourceName, "state_info.aws_transit_gateway_attachment_id"),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentResourceName, "peering_connection_id"),
		testAccCheckAivenTransitGatewayVPCAttachmentIDParts(idParts),

		testAccCheckAivenAWSTransitGatewayVPCAttachmentDataSourceID(c.Region),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "vpc_id", transitGatewayAttachmentResourceName, "vpc_id"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_cloud_account", transitGatewayAttachmentResourceName, "peer_cloud_account"),
		resource.TestCheckResourceAttrPair(transitGatewayAttachmentDataSourceName, "peer_vpc", transitGatewayAttachmentResourceName, "peer_vpc"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "peer_region", c.Region),
		resource.TestMatchResourceAttr(transitGatewayAttachmentDataSourceName, "state", liveAttachmentStatePattern),
		resource.TestCheckResourceAttrPair(
			transitGatewayAttachmentDataSourceName,
			"state_info.aws_transit_gateway_attachment_id",
			transitGatewayAttachmentResourceName,
			"state_info.aws_transit_gateway_attachment_id",
		),
		resource.TestCheckNoResourceAttr(transitGatewayAttachmentDataSourceName, "peering_connection_id"),
		resource.TestCheckResourceAttr(transitGatewayAttachmentDataSourceName, "user_peer_network_cidrs.#", strconv.Itoa(len(cidrs))),
	}
	for _, cidr := range cidrs {
		checks = append(checks,
			resource.TestCheckTypeSetElemAttr(transitGatewayAttachmentResourceName, "user_peer_network_cidrs.*", cidr),
			resource.TestCheckTypeSetElemAttr(transitGatewayAttachmentDataSourceName, "user_peer_network_cidrs.*", cidr),
		)
	}

	return resource.ComposeTestCheckFunc(checks...)
}

func testAccCheckAivenAWSTransitGatewayVPCAttachmentDataSourceID(peerRegion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attachment, ok := s.RootModule().Resources[transitGatewayAttachmentResourceName]
		if !ok {
			return fmt.Errorf("transit gateway VPC attachment not found in state")
		}
		dataSource, ok := s.RootModule().Resources[transitGatewayAttachmentDataSourceName]
		if !ok {
			return fmt.Errorf("transit gateway VPC attachment data source not found in state")
		}

		want := attachment.Primary.ID
		switch len(strings.Split(want, "/")) {
		case 4:
			want += "/" + peerRegion
		case 5:
		default:
			return fmt.Errorf("expected resource ID with 4-5 parts, got %q", attachment.Primary.ID)
		}
		if got := dataSource.Primary.ID; got != want {
			return fmt.Errorf(
				"expected transit gateway VPC attachment data source ID %q, got %q",
				want,
				got,
			)
		}
		return nil
	}
}

func testAccCheckAivenAWSTransitGatewayVPCAttachmentActive(ctx context.Context, peerRegion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attachment, ok := s.RootModule().Resources[transitGatewayAttachmentResourceName]
		if !ok {
			return fmt.Errorf("transit gateway VPC attachment not found in state")
		}

		parts := strings.Split(attachment.Primary.ID, "/")
		if len(parts) != 4 && len(parts) != 5 {
			return fmt.Errorf("parse transit gateway VPC attachment ID: expected 4-5 parts, got %d", len(parts))
		}
		project, projectVPCID, peerCloudAccount, peerVPC := parts[0], parts[1], parts[2], parts[3]
		if len(parts) == 5 && parts[4] != peerRegion {
			return fmt.Errorf("expected peer region %q in transit gateway VPC attachment ID, got %q", peerRegion, parts[4])
		}

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
					if connection.PeerRegion == nil || *connection.PeerRegion != peerRegion {
						continue
					}
					if connection.VpcPeeringConnectionType != vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment {
						return nil, "", fmt.Errorf("expected %q peering connection type, got %q", vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment, connection.VpcPeeringConnectionType)
					}

					return &connection, string(connection.State), nil
				}

				return nil, "", fmt.Errorf("transit gateway VPC attachment %q not found in Aiven VPC", attachment.Primary.ID)
			},
			Timeout:      10 * time.Minute,
			PollInterval: 5 * time.Second,
		}
		if _, err := wait.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("wait for active transit gateway VPC attachment: %w", err)
		}

		return nil
	}
}

func testAccCheckAivenAWSTransitGatewayVPCAttachmentDeleted(ctx context.Context, peerResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		projectVPC, ok := s.RootModule().Resources["aiven_project_vpc.aiven_vpc"]
		if !ok {
			return fmt.Errorf("Aiven project VPC not found in state")
		}
		peer, ok := s.RootModule().Resources[peerResourceName]
		if !ok {
			return fmt.Errorf("peer AWS resource %q not found in state", peerResourceName)
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

		peerCloudAccount := peer.Primary.Attributes["owner_id"]
		for _, connection := range vpcOut.PeeringConnections {
			if connection.PeerCloudAccount == peerCloudAccount && connection.PeerVpc == peer.Primary.ID {
				return fmt.Errorf("VPC peering connection for AWS resource %q still exists in state %q", peer.Primary.ID, connection.State)
			}
		}

		return nil
	}
}

func testAccAivenAWSTransitGatewayVPCAttachment(prefix string, c *awsConfig, cidrs *[]string, peerRegion *string) string {
	peerRegionValue := "null"
	if peerRegion != nil {
		peerRegionValue = strconv.Quote(*peerRegion)
	}
	cidrsValue := "null"
	if cidrs != nil {
		cidrsValue = formatCIDRs(*cidrs)
	}

	return testAccAivenAWSTransitGatewayVPCAttachmentFixture(prefix, c) + fmt.Sprintf(`
resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.aiven_vpc.id
  peer_cloud_account = aws_ec2_transit_gateway.transit_gateway.owner_id
  peer_region        = %[1]s
  peer_vpc           = aws_ec2_transit_gateway.transit_gateway.id

  user_peer_network_cidrs = %[2]s

  depends_on = [
    aws_ram_resource_association.transit_gateway,
    aws_ram_principal_association.aiven,
  ]

  timeouts {
    create = "10m"
  }
}

data "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_transit_gateway_vpc_attachment.attachment.vpc_id
  peer_cloud_account = aiven_transit_gateway_vpc_attachment.attachment.peer_cloud_account
  peer_vpc           = aiven_transit_gateway_vpc_attachment.attachment.peer_vpc
}
`, peerRegionValue, cidrsValue)
}

func testAccAivenAWSTransitGatewayVPCAttachmentFixture(prefix string, c *awsConfig) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "aws" {
  region = %[3]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-%[3]s"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aws_ec2_transit_gateway" "transit_gateway" {
  auto_accept_shared_attachments = "enable"

  tags = {
    Name = "%[1]s-transit-gateway"
  }

  timeouts {
    delete = "20m"
  }
}

resource "aws_ram_resource_share" "transit_gateway" {
  name                      = "%[1]s-transit-gateway"
  allow_external_principals = true
}

resource "aws_ram_resource_association" "transit_gateway" {
  resource_arn       = aws_ec2_transit_gateway.transit_gateway.arn
  resource_share_arn = aws_ram_resource_share.transit_gateway.arn
}

resource "aws_ram_principal_association" "aiven" {
  principal          = %[4]q
  resource_share_arn = aws_ram_resource_share.transit_gateway.arn
}
`, prefix, c.Project, c.Region, aivenAWSAccountID)
}

func testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeering(prefix string, c *awsConfig) string {
	return testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeeringFixture(prefix, c) + fmt.Sprintf(`
resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.aiven_vpc.id
  peer_cloud_account = aws_vpc.peer.owner_id
  peer_region        = %[1]q
  peer_vpc           = aws_vpc.peer.id

  user_peer_network_cidrs = []

  timeouts {
    create = "10m"
  }
}

data "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_transit_gateway_vpc_attachment.attachment.vpc_id
  peer_cloud_account = aiven_transit_gateway_vpc_attachment.attachment.peer_cloud_account
  peer_vpc           = aiven_transit_gateway_vpc_attachment.attachment.peer_vpc
}
`, c.Region)
}

func testAccAivenAWSTransitGatewayVPCAttachmentLegacyVPCPeeringFixture(prefix string, c *awsConfig) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "aws" {
  region = %[3]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-%[3]s"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "%[1]s-legacy-vpc-peering"
  }
}
`, prefix, c.Project, c.Region)
}

func formatCIDRs(cidrs []string) string {
	if len(cidrs) == 0 {
		return "[]"
	}

	quoted := make([]string, len(cidrs))
	for i, cidr := range cidrs {
		quoted[i] = strconv.Quote(cidr)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func getAWSConfig(t *testing.T) *awsConfig {
	t.Helper()

	if os.Getenv(awsTransitGatewayTestEnv) == "" {
		t.Skipf("environment variable %s must be set to run this test", awsTransitGatewayTestEnv)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = defaultAWSRegion
	}

	return &awsConfig{
		Project: acc.ProjectName(),
		Region:  region,
	}
}

func awsExternalProvider() resource.ExternalProvider {
	return resource.ExternalProvider{
		Source:            "hashicorp/aws",
		VersionConstraint: "=6.23.0",
	}
}
