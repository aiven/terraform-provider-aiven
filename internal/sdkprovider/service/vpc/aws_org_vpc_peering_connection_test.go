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
		orgName        = acc.SkipIfEnvVarsNotSet(t, "AIVEN_ORGANIZATION_NAME")["AIVEN_ORGANIZATION_NAME"]
		registry       = preSetAwsOrgVPCPeeringTemplates(t)
		newComposition = func() *acc.CompositionBuilder {
			return registry.NewCompositionBuilder().
				Add("organization_data", map[string]any{
					"organization_name": orgName})
		}

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
				Config: newComposition().
					Add(organizationVPCResource, map[string]any{
						"resource_name": "test_org_vpc",
						"cloud_name":    fmt.Sprintf("aws-%s", awsRegion),
						"network_cidr":  "10.0.0.0/24",
					}).
					Add(awsOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_org_vpc_peering",
						"organization_id":     acc.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": acc.Reference("aiven_organization_vpc.test_org_vpc.organization_vpc_id"),
						"aws_account_id":      acc.Literal(awsAccountID),
						"aws_vpc_id":          acc.Literal(awsVpcID),
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
		orgName        = envVars["AIVEN_ORGANIZATION_NAME"]
		awsRegion      = "eu-central-1"
		registry       = preSetAwsOrgVPCPeeringTemplates(t)
		newComposition = func() *acc.CompositionBuilder {
			return registry.NewCompositionBuilder().
				Add("aws_provider", map[string]any{
					"aws_region": awsRegion}).
				Add("organization_data", map[string]any{
					"organization_name": orgName})
		}

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
				Config: newComposition().
					Add("aws_vpc", map[string]any{
						"resource_name": "example",
						"cidr_block":    "172.16.0.0/16",
						"vpc_name":      fmt.Sprintf("%s-vpc", serviceName),
					}).
					Add("aws_route_table", map[string]any{
						"resource_name":    "example",
						"vpc_id":           "aws_vpc.example.id",
						"route_table_name": fmt.Sprintf("%s-route-table", serviceName),
					}).
					Add(organizationVPCResource, map[string]any{
						"resource_name": "example",
						"cloud_name":    fmt.Sprintf("aws-%s", awsRegion),
						"network_cidr":  "10.0.0.0/24",
					}).
					Add(awsOrgVPCPeeringResource, map[string]any{
						"resource_name":       "test_peering",
						"organization_id":     acc.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": acc.Reference("aiven_organization_vpc.example.organization_vpc_id"),
						"aws_account_id":      acc.Reference("aws_vpc.example.owner_id"),
						"aws_vpc_id":          acc.Reference("aws_vpc.example.id"),
						"aws_vpc_region":      awsRegion,
					}).
					Add("peering_datasource", map[string]any{
						"resource_name":       "test_peering",
						"organization_id":     acc.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": acc.Reference("aiven_organization_vpc.example.organization_vpc_id"),
						"aws_account_id":      acc.Reference("aws_vpc.example.owner_id"),
						"aws_vpc_id":          acc.Reference("aws_vpc.example.id"),
						"aws_vpc_region":      awsRegion,
					}).
					Add("aws_vpc_peering_accepter", map[string]any{
						"resource_name":         "example",
						"peering_connection_id": "aiven_aws_org_vpc_peering_connection.test_peering.aws_vpc_peering_connection_id",
						"peering_name":          fmt.Sprintf("%s-peering-accepter", serviceName),
					}).
					Add("aws_route", map[string]any{
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

func preSetAwsOrgVPCPeeringTemplates(t *testing.T) *acc.TemplateRegistry {
	t.Helper()

	registry := acc.NewTemplateRegistry(awsOrgVPCPeeringResource)

	registry.MustAddTemplate(t, "organization_data", `
data "aiven_organization" "foo" {
  name = "{{ .organization_name }}"
}`)

	registry.MustAddTemplate(t, "aws_provider", `
provider "aws" {
	region = "{{ .aws_region }}"
}`)

	registry.MustAddTemplate(t, organizationVPCResource, `
resource "aiven_organization_vpc" "{{ .resource_name }}" {
    organization_id = data.aiven_organization.foo.id
    cloud_name     = "{{ .cloud_name }}"
    network_cidr   = "{{ .network_cidr }}"
}`)

	registry.MustAddTemplate(t, awsOrgVPCPeeringResource, `
resource "aiven_aws_org_vpc_peering_connection" "{{ .resource_name }}" {
    organization_id         	= {{ if .organization_id.IsLiteral }}"{{ .organization_id.Value }}"{{ else }}{{ .organization_id.Value }}{{ end }}
    organization_vpc_id         = {{ if .organization_vpc_id.IsLiteral }}"{{ .organization_vpc_id.Value }}"{{ else }}{{ .organization_vpc_id.Value }}{{ end }}
    aws_account_id 	= {{ if .aws_account_id.IsLiteral }}"{{ .aws_account_id.Value }}"{{ else }}{{ .aws_account_id.Value }}{{ end }}
    aws_vpc_id 		= {{ if .aws_vpc_id.IsLiteral }}"{{ .aws_vpc_id.Value }}"{{ else }}{{ .aws_vpc_id.Value }}{{ end }}
    aws_vpc_region 	= "{{ .aws_vpc_region }}"
}`)

	registry.MustAddTemplate(t, "peering_datasource", `
data "aiven_aws_org_vpc_peering_connection" "{{ .resource_name }}" {
    organization_id         	= {{ if .organization_id.IsLiteral }}"{{ .organization_id.Value }}"{{ else }}{{ .organization_id.Value }}{{ end }}
    organization_vpc_id         = {{ if .organization_vpc_id.IsLiteral }}"{{ .organization_vpc_id.Value }}"{{ else }}{{ .organization_vpc_id.Value }}{{ end }}
    aws_account_id 	= {{ if .aws_account_id.IsLiteral }}"{{ .aws_account_id.Value }}"{{ else }}{{ .aws_account_id.Value }}{{ end }}
    aws_vpc_id 		= {{ if .aws_vpc_id.IsLiteral }}"{{ .aws_vpc_id.Value }}"{{ else }}{{ .aws_vpc_id.Value }}{{ end }}
    aws_vpc_region 	= "{{ .aws_vpc_region }}"
}`)

	// AWS VPC Resource
	registry.MustAddTemplate(t, "aws_vpc", `
resource "aws_vpc" "{{ .resource_name }}" {
    cidr_block           = "{{ .cidr_block }}"
    enable_dns_hostnames = true
    enable_dns_support   = true

    tags = {
        Name = "{{ .vpc_name }}"
    }
}`)

	// AWS Route Table
	registry.MustAddTemplate(t, "aws_route_table", `
resource "aws_route_table" "{{ .resource_name }}" {
    vpc_id = {{ .vpc_id }}

    tags = {
        Name = "{{ .route_table_name }}"
    }
}`)

	// AWS VPC Peering Connection Accepter
	registry.MustAddTemplate(t, "aws_vpc_peering_accepter", `
resource "aws_vpc_peering_connection_accepter" "{{ .resource_name }}" {
    vpc_peering_connection_id = {{ .peering_connection_id }}
    auto_accept              = true

    tags = {
        Name = "{{ .peering_name }}"
    }
}`)

	// AWS Route for Aiven VPC CIDR
	registry.MustAddTemplate(t, "aws_route", `
resource "aws_route" "{{ .resource_name }}" {
    route_table_id            = {{ .route_table_id }}
    destination_cidr_block    = {{ .destination_cidr }}
    vpc_peering_connection_id = {{ .peering_connection_id }}
}`)

	return registry
}
