package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenAWSPrivatelink_basic(t *testing.T) {
	var s awsSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	prefix := "test-tf-acc-" + acctest.RandString(7)
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
				Config: testAccAWSPrivatelinkResource(prefix, &s),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("aiven_pg.pg", "state", "RUNNING"),
					resource.TestCheckResourceAttr("aiven_aws_privatelink.privatelink", "service_name", prefix+"-pg"),
					resource.TestCheckResourceAttrSet("aiven_aws_privatelink.privatelink", "aws_service_name"),
					resource.TestCheckResourceAttrSet("aiven_aws_privatelink.privatelink", "aws_service_id"),

					// Aws resources
					resource.TestCheckResourceAttrSet("aws_vpc.aws_vpc", "id"),
					resource.TestCheckResourceAttrSet("aws_vpc_endpoint.endpoint", "id"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.endpoint", "state", "available"),
				),
			},
			importStateByName("aiven_project_vpc.aiven_vpc"),
			importStateByName("aiven_pg.pg"),
			importStateByName("aiven_aws_privatelink.privatelink"),
			importStateByName("aws_vpc.aws_vpc"),
			importStateByName("aws_vpc_endpoint.endpoint"),
		},
	})
}

func testAccAWSPrivatelinkResource(prefix string, s *awsSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "aws" {
  region = "eu-west-3"
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-eu-west-3"
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
`, prefix, s.Project, s.Principal)
}
