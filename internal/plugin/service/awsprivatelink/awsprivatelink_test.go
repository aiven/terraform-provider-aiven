package awsprivatelink_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

type awsPrivatelinkSecrets struct {
	Project   string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	Principal string `envconfig:"AWS_PRIVATE_LINK_PRINCIPAL" required:"true"`
	Region    string `envconfig:"AWS_DEFAULT_REGION" default:"eu-west-3"`
}

func TestAccAivenAWSPrivatelink_basic(t *testing.T) {
	var s awsPrivatelinkSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	prefix := "test-tf-acc-" + acctest.RandString(7)
	resourceName := "aiven_aws_privatelink.privatelink"
	dataSourceName := "data.aiven_aws_privatelink.privatelink"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {
				Source:            "hashicorp/aws",
				VersionConstraint: "=4.40.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPrivatelinkResource(prefix, s.Project, s.Principal, s.Region, true),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("aiven_pg.pg", "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_name", prefix+"-pg"),
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_service_name"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_service_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "project", resourceName, "project"),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service_name", resourceName, "aws_service_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service_id", resourceName, "aws_service_id"),
					resource.TestCheckResourceAttr(dataSourceName, "state", "active"),

					// AWS resources
					resource.TestCheckResourceAttrSet("aws_vpc.aws_vpc", "id"),
					resource.TestCheckResourceAttrSet("aws_vpc_endpoint.endpoint", "id"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.endpoint", "state", "available"),
				),
			},
			importStateByName("aiven_project_vpc.aiven_vpc"),
			importStateByName("aiven_pg.pg"),
			importStateByName(resourceName),
			importStateByName("aws_vpc.aws_vpc"),
			importStateByName("aws_vpc_endpoint.endpoint"),
		},
	})
}

func TestAccAivenAWSPrivatelink_backwardCompat(t *testing.T) {
	principal := requireAWSPrivatelinkPrincipal(t)
	projectName := acc.ProjectName()
	prefix := "test-tf-acc-" + acctest.RandString(7)
	resourceName := "aiven_aws_privatelink.privatelink"
	config := testAccAWSPrivatelinkResource(prefix, projectName, principal, "eu-west-3", false)

	oldChecks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
		resource.TestCheckResourceAttr("aiven_pg.pg", "state", "RUNNING"),
		resource.TestCheckResourceAttr(resourceName, "project", projectName),
		resource.TestCheckResourceAttr(resourceName, "service_name", prefix+"-pg"),
		resource.TestCheckResourceAttrSet(resourceName, "aws_service_name"),
		resource.TestCheckResourceAttrSet(resourceName, "aws_service_id"),
	)
	newChecks := resource.ComposeTestCheckFunc(
		oldChecks,
		resource.TestCheckResourceAttr(resourceName, "state", "active"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aiven": acc.ExternalAivenProvider(t, "4.56.0"),
				},
				Config: config,
				Check:  oldChecks,
			},
			{
				ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
				Config:                   config,
				ExpectNonEmptyPlan:       false,
				Check:                    newChecks,
			},
		},
	})
}

func testAccAWSPrivatelinkResource(prefix, project, principal, region string, includeAWSEndpoint bool) string {
	awsProvider := ""
	awsEndpoint := ""
	if includeAWSEndpoint {
		awsProvider = fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}
`, region)
		awsEndpoint = fmt.Sprintf(`
resource "aws_vpc" "aws_vpc" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "%[1]s-privatelink"
  }
}

resource "aws_vpc_endpoint" "endpoint" {
  vpc_id            = aws_vpc.aws_vpc.id
  service_name      = aiven_aws_privatelink.privatelink.aws_service_name
  vpc_endpoint_type = "Interface"
}
`, prefix)
	}

	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}
%[6]s

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-%[4]s"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aiven_pg" "pg" {
  project        = aiven_project_vpc.aiven_vpc.project
  cloud_name     = aiven_project_vpc.aiven_vpc.cloud_name
  project_vpc_id = aiven_project_vpc.aiven_vpc.id
  plan           = "startup-4"
  service_name   = "%[1]s-pg"
}

resource "aiven_aws_privatelink" "privatelink" {
  project      = data.aiven_project.project.project
  service_name = aiven_pg.pg.service_name

  principals = [
    %[3]q
  ]
}

data "aiven_aws_privatelink" "privatelink" {
  project      = aiven_aws_privatelink.privatelink.project
  service_name = aiven_aws_privatelink.privatelink.service_name

  depends_on = [aiven_aws_privatelink.privatelink]
}
%[5]s`, prefix, project, principal, region, awsEndpoint, awsProvider)
}

func importStateByName(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName: name,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			root := s.RootModule()
			rs, ok := root.Resources[name]
			if !ok {
				return "", fmt.Errorf(`resource %q not found in the state`, name)
			}
			return rs.Primary.ID, nil
		},
	}
}

func requireAWSPrivatelinkPrincipal(t *testing.T) string {
	t.Helper()

	principal, ok := os.LookupEnv("AWS_PRIVATE_LINK_PRINCIPAL")
	if !ok {
		t.Skip("environment variable AWS_PRIVATE_LINK_PRINCIPAL not set")
	}
	return principal
}
