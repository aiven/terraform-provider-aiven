package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

type awsSecrets struct {
	Project   string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	AccountID string `envconfig:"AWS_ACCOUNT_ID" required:"true"`
	Region    string `envconfig:"AWS_REGION" required:"true"`

	// Don't need to pass to tf file, this must be in env
	AccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	SessionToken    string `envconfig:"AWS_SESSION_TOKEN"`
}

func TestAccAivenAWSVPCPeeringConnection_e2e(t *testing.T) {
	var s awsSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

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
				Config: testAccAivenAWSVPCPeeringConnectionTF(&s),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),

					// We can't check peering_connection state, because it's updated async and gets ACTIVE any time later
					resource.TestCheckResourceAttrSet("aiven_aws_vpc_peering_connection.peering_connection", "id"),
					resource.TestCheckResourceAttr("aiven_aws_vpc_peering_connection.peering_connection", "aws_account_id", s.AccountID),
					resource.TestCheckResourceAttr("aiven_aws_vpc_peering_connection.peering_connection", "aws_vpc_region", s.Region),

					// Azure resources
					resource.TestCheckResourceAttrSet("aws_vpc.aws_vpc", "id"),
					resource.TestCheckResourceAttr("aws_vpc.aws_vpc", "cidr_block", "10.0.0.0/24"),
				),
			},
			importStateByName("aiven_project_vpc.aiven_vpc"),
			importStateByName("aiven_aws_vpc_peering_connection.peering_connection"),
			importStateByName("aws_vpc.aws_vpc"),
		},
	})
}

func testAccAivenAWSVPCPeeringConnectionTF(s *awsSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[1]q
}

provider "aws" {
  region = %[2]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "aws-eu-west-1"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aws_vpc" "aws_vpc" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "test-acc-tf-vpc-peering"
  }
}

resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = aiven_project_vpc.aiven_vpc.id
  aws_account_id = %[3]q
  aws_vpc_region = %[2]q
  aws_vpc_id     = aws_vpc.aws_vpc.id
}
`, s.Project, s.Region, s.AccountID)
}
