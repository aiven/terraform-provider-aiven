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
	awsOrgVPCPeeringResource = "aiven_aws_org_vpc_peering_connection"
)

// TestAccAivenAWSOrgVPCPeeringConnection tests the AWS VPC peering connection resource functionality.
// Since creating a real AWS VPC peering connection in CI is not feasible now, this test:
// 1. Sets up a test environment with a fake AWS account ID and VPC ID
// 2. Attempts to create an Aiven VPC and a peering connection
// 3. Validates that the creation fails with the expected error due to invalid AWS credentials
func TestAccAivenAWSOrgVPCPeeringConnection(t *testing.T) {
	var (
		orgName = acc.SkipIfEnvVarsNotSet(t, "AIVEN_ORGANIZATION_NAME")["AIVEN_ORGANIZATION_NAME"]

		templBuilder = template.InitializeTemplateStore(t).NewBuilder().
				AddDataSource("aiven_organization", map[string]interface{}{
				"resource_name": "foo",
				"name":          orgName,
			})

		awsAccountID = "123456789012"          // Fake AWS account ID
		awsVpcID     = "vpc-1a1a111a111a11a11" // Fake AWS VPC ID
		awsRegion    = "eu-west-2"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAWSOrgVPCPeeringResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      fmt.Sprintf("aws-%s", awsRegion),
						"network_cidr":    "10.0.0.0/24",
					}).
					AddResource(awsOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_org_vpc_peering",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference("aiven_organization_vpc.test_org_vpc.organization_vpc_id"),
						"aws_account_id":      template.Literal(awsAccountID),
						"aws_vpc_id":          template.Literal(awsVpcID),
						"aws_vpc_region":      awsRegion,
					}).MustRender(t),
				ExpectError: regexp.MustCompile(`VPC peering connection cannot be created`), // Expected error due to invalid AWS account ID
			},
		},
	})
}

// TestAccAivenAWSOrgVPCPeeringConnectionFull tests the complete AWS VPC peering connection workflow
// with real AWS resources. This test:
// 1. Creates an AWS VPC and route table
// 2. Creates an Aiven Organization VPC
// 3. Establishes VPC peering between AWS and Aiven VPCs
// 4. Accepts the peering connection on AWS side
// 5. Sets up routing for the peered VPCs
//
// Note: The test will be skipped in CI environments since it requires real AWS credentials
// and resources. This test is meant for local development and verification for now.
// Prerequisites:
// - Valid AWS credentials with permissions to create/delete VPC resources
// - Proper AWS profile configuration (can be set via AWS_PROFILE env var)
// - Required permissions: VPC creation/deletion, VPC peering, route table management
func TestAccAivenAWSOrgVPCPeeringConnectionFull(t *testing.T) {
	var envVars = acc.SkipIfEnvVarsNotSet(
		t,
		"AIVEN_ORGANIZATION_NAME",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
	)

	var (
		orgName   = envVars["AIVEN_ORGANIZATION_NAME"]
		awsRegion = "eu-central-1"

		templBuilder = template.InitializeTemplateStore(t).NewBuilder().
				AddDataSource("aiven_organization", map[string]any{
				"resource_name": "foo",
				"name":          orgName,
			})

		resourceName = fmt.Sprintf("%s.%s", awsOrgVPCPeeringResource, "test_peering")

		randName    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-acc-%s", randName)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckAWSOrgVPCPeeringResourceDestroy,
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {
				Source:            "hashicorp/aws",
				VersionConstraint: "=5.84.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddResource("aws_vpc", map[string]any{
						"resource_name": "example",
						"cidr_block":    "172.16.0.0/16",
						"vpc_name":      fmt.Sprintf("%s-vpc", serviceName),
					}).
					AddResource("aws_route_table", map[string]any{
						"resource_name":    "example",
						"vpc_id":           "aws_vpc.example.id",
						"route_table_name": fmt.Sprintf("%s-route-table", serviceName),
					}).
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "example",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      fmt.Sprintf("aws-%s", awsRegion),
						"network_cidr":    "10.0.0.0/24",
					}).
					AddResource(awsOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_peering",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference("aiven_organization_vpc.example.organization_vpc_id"),
						"aws_account_id":      template.Reference("aws_vpc.example.owner_id"),
						"aws_vpc_id":          template.Reference("aws_vpc.example.id"),
						"aws_vpc_region":      awsRegion,
					}).
					AddDataSource(awsOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_peering",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference("aiven_organization_vpc.example.organization_vpc_id"),
						"aws_account_id":      template.Reference("aws_vpc.example.owner_id"),
						"aws_vpc_id":          template.Reference("aws_vpc.example.id"),
						"aws_vpc_region":      awsRegion,
					}).
					AddResource("aws_vpc_peering_accepter", map[string]any{
						"resource_name":         "example",
						"peering_connection_id": "aiven_aws_org_vpc_peering_connection.test_peering.aws_vpc_peering_connection_id",
						"peering_name":          fmt.Sprintf("%s-peering-accepter", serviceName),
					}).
					AddResource("aws_route", map[string]any{
						"resource_name":         "aiven_vpc_route",
						"route_table_id":        "aws_route_table.example.id",
						"destination_cidr":      "aiven_organization_vpc.example.network_cidr",
						"peering_connection_id": "aws_vpc_peering_connection_accepter.example.vpc_peering_connection_id",
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_vpc_id", "aiven_organization_vpc.example", "organization_vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_account_id", "aws_vpc.example", "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_vpc_id", "aws_vpc.example", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.foo", "id"),
					resource.TestCheckResourceAttr(resourceName, "aws_vpc_region", awsRegion),
					resource.TestCheckResourceAttrSet(resourceName, "peering_connection_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_vpc_peering_connection_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),

					resource.TestCheckResourceAttrPair(fmt.Sprintf("data.%s.%s", awsOrgVPCPeeringResource, "test_peering"), "organization_id", "data.aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(fmt.Sprintf("data.%s.%s", awsOrgVPCPeeringResource, "test_peering"), "organization_vpc_id", "aiven_organization_vpc.example", "organization_vpc_id"),
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

func testAccCheckAWSOrgVPCPeeringResourceDestroy(s *terraform.State) error {
	ctx := context.Background()

	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error initializing Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != awsOrgVPCPeeringResource {
			continue
		}

		orgID, orgVpcID, awsAccountID, awsVpcID, awsRegion, err := schemautil.SplitResourceID5(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting resource with ID: %q - %w", rs.Primary.ID, err)
		}

		orgVPC, err := c.OrganizationVpcGet(ctx, orgID, orgVpcID)
		if common.IsCritical(err) {
			return fmt.Errorf("error fetching VPC (%q): %w", orgVpcID, err)
		}

		if orgVPC == nil {
			return nil // Peering connection was deleted with the VPC
		}

		var pc *organizationvpc.OrganizationVpcGetPeeringConnectionOut
		for _, pCon := range orgVPC.PeeringConnections {
			if pCon.PeerCloudAccount == awsAccountID &&
				pCon.PeerVpc == awsVpcID &&
				pCon.PeerRegion != nil &&
				*pCon.PeerRegion == awsRegion &&
				pCon.PeeringConnectionId != nil {
				pc = &pCon

				break
			}
		}

		if pc != nil {
			return fmt.Errorf("peering connection %q still exists", *pc.PeeringConnectionId)
		}
	}

	return nil
}
