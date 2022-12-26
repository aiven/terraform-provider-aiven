package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenAWSTransitGatewayVPCAttachment_basic(t *testing.T) {
	var s awsSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	prefix := "test-tf-acc-" + acctest.RandString(7)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {
				Source:            "hashicorp/aws",
				VersionConstraint: "=4.40.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAivenAWSTransitGatewayVPCAttachment(prefix, &s),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("aiven_transit_gateway_vpc_attachment.attachment", "peer_cloud_account", s.AccountID),
					resource.TestCheckResourceAttr("aiven_transit_gateway_vpc_attachment.attachment", "peer_region", s.Region),
					resource.TestCheckResourceAttr("aiven_transit_gateway_vpc_attachment.attachment", "user_peer_network_cidrs.#", "0"),
					resource.TestCheckResourceAttrSet("aiven_transit_gateway_vpc_attachment.attachment", "state"),

					// Azure resources
					resource.TestCheckResourceAttrSet("aws_vpc.aws_vpc", "id"),
					resource.TestCheckResourceAttr("aws_vpc.aws_vpc", "cidr_block", "10.0.0.0/24"),
				),
			},
			importStateByName("aiven_project_vpc.aiven_vpc"),
			importStateByName("aiven_transit_gateway_vpc_attachment.attachment"),
			importStateByName("aws_vpc.aws_vpc"),
		},
	})
}

func testAccAivenAWSTransitGatewayVPCAttachment(prefix string, s *awsSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "aws" {
  region = %[3]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-eu-west-2"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aws_vpc" "aws_vpc" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "%[1]s-transit-gateway"
  }
}

resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.aiven_vpc.id
  peer_cloud_account = %[4]q
  peer_region        = %[3]q
  peer_vpc           = aws_vpc.aws_vpc.id

  user_peer_network_cidrs = [
  ]

  timeouts {
    create = "10m"
  }
}
`, prefix, s.Project, s.Region, s.AccountID)
}
