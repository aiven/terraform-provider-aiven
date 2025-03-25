package vpc_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	gcpOrgVPCPeeringResource = "aiven_gcp_org_vpc_peering_connection"
)

// TestAccAivenGCPOrgVPCPeeringConnection tests the GCP VPC peering connection resource functionality.
// Since creating a real GCP VPC peering connection in CI requires valid GCP credentials, this test:
// 1. Sets up a test environment with an invalid GCP project ID and VPC ID
// 2. Attempts to create an Aiven VPC and a peering connection
// 3. Validates that the creation fails with the expected error due to invalid GCP project ID
func TestAccAivenGCPOrgVPCPeeringConnection(t *testing.T) {
	var (
		orgName      = acc.OrganizationName()
		templBuilder = template.InitializeTemplateStore(t).NewBuilder()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGCPOrgVPCPeeringResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddDataSource("aiven_organization", map[string]any{
						"resource_name": "foo",
						"name":          orgName,
					}).
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "google-europe-west10",
						"network_cidr":    "10.0.0.0/24",
					}).
					AddResource(gcpOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_org_vpc_peering",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference("aiven_organization_vpc.test_org_vpc.organization_vpc_id"),
						"gcp_project_id":      template.Literal("wrong_project_id"),
						"peer_vpc":            template.Literal("wrong_peer_vpc"),
					}).MustRender(t),
				ExpectError: regexp.MustCompile(`peer_cloud_account must be a valid GCP project ID`), // Expected error due to invalid GCP arguments
			},
		},
	})
}

// TestAccAivenGCPOrgVPCPeeringConnectionFull tests the complete GCP VPC peering connection workflow
// with real GCP resources. This test:
// 1. Creates a GCP VPC and route
// 2. Creates an Aiven Organization VPC
// 3. Establishes VPC peering between GCP and Aiven VPCs
// 4. Sets up routing for the peered VPCs
//
// Note: The test will be skipped in CI environments since it requires real GCP credentials
// and resources. This test is meant for local development and verification for now.
// Prerequisites:
// - Valid GCP credentials with permissions to create/delete VPC resources
// - GCP project with VPC API enabled
// - Required permissions: VPC creation/deletion, VPC peering, route management
func TestAccAivenGCPOrgVPCPeeringConnectionFull(t *testing.T) {
	var (
		orgName   = acc.OrganizationName()
		gcpRegion = "europe-west10"
	)

	var (
		gcpProject   = acc.RequireEnvVars(t, "GOOGLE_PROJECT")["GOOGLE_PROJECT"]
		resourceName = fmt.Sprintf("%s.%s", gcpOrgVPCPeeringResource, "test_peering")

		randName    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-acc-%s", randName)

		// Register the templates needed for this test
		templBuilder = template.InitializeTemplateStore(t).NewBuilder()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckGCPOrgVPCPeeringResourceDestroy,
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": {
				Source:            "hashicorp/google",
				VersionConstraint: "=6.15.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddTemplate("gcp_provider", map[string]any{
						"gcp_project": gcpProject,
						"gcp_region":  gcpRegion,
					}).
					AddDataSource("aiven_organization", map[string]any{
						"resource_name": "foo",
						"name":          orgName,
					}).
					AddResource("google_compute_network", map[string]any{
						"resource_name":           "example",
						"network_name":            fmt.Sprintf("%s-vpc", serviceName),
						"auto_create_subnetworks": false,
					}).
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "example",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      fmt.Sprintf("google-%s", gcpRegion),
						"network_cidr":    "10.0.0.0/24",
					}).
					AddResource(gcpOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_peering",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference("aiven_organization_vpc.example.organization_vpc_id"),
						"gcp_project_id":      template.Literal(gcpProject),
						"peer_vpc":            template.Reference("google_compute_network.example.name"),
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_vpc_id", "aiven_organization_vpc.example", "organization_vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "gcp_project_id", gcpProject),
					resource.TestCheckResourceAttrPair(resourceName, "peer_vpc", "google_compute_network.example", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "self_link"),
				),
			},
			{
				// importing the resource
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGCPOrgVPCPeeringResourceDestroy(s *terraform.State) error {
	ctx := context.Background()

	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error initializing Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != gcpOrgVPCPeeringResource {
			continue
		}

		orgID, vpcID, cloudAcc, peerVPC, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing GCP peering VPC ID: %q. %w", rs.Primary.ID, err)
		}

		orgVPC, err := c.OrganizationVpcGet(ctx, orgID, vpcID)
		if common.IsCritical(err) {
			return fmt.Errorf("error fetching VPC (%q): %w", orgID, err)
		}

		if orgVPC == nil {
			return nil // Peering connection was deleted with the VPC
		}

		var pc *organizationvpc.OrganizationVpcGetPeeringConnectionOut
		for _, p := range orgVPC.PeeringConnections {
			if p.PeerCloudAccount == cloudAcc && p.PeerVpc == peerVPC && p.PeeringConnectionId != nil {
				pc = &p
				break
			}
		}

		if pc != nil {
			return fmt.Errorf("peering connection %q still exists", *pc.PeeringConnectionId)
		}
	}

	return nil
}
