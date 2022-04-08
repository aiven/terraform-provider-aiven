package vpc_test

import (
	"fmt"
	"os"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenAWSVPCPeeringConnection_basic(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" ||
		os.Getenv("AWS_VPC_ID") == "" ||
		os.Getenv("AWS_ACCOUNT_ID") == "" {
		t.Skip("env variables AWS_REGION, AWS_VPC_ID and AWS_ACCOUNT_ID required to run this test")
	}

	resourceName := "aiven_vpc_peering_connection.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAWSResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "peer_cloud_account", os.Getenv("AWS_ACCOUNT_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_vpc", os.Getenv("AWS_VPC_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_region", os.Getenv("AWS_REGION")),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionAWSResource() string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_project_vpc" "bar" {
		  project      = data.aiven_project.foo.project
		  cloud_name   = "aws-%s"
		  network_cidr = "10.0.0.0/24"
		
		  timeouts {
		    create = "5m"
		  }
		}
		
		resource "aiven_aws_vpc_peering_connection" "foo" {
		  vpc_id             = aiven_project_vpc.bar.id
		  aws_account_id = "%s"
		  aws_vps_id     = "%s"
		  aws_vpc_region = "%s"
		
		  timeouts {
		    create = "10m"
		  }
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ACCOUNT_ID"),
		os.Getenv("AWS_VPC_ID"),
		os.Getenv("AWS_REGION"))
}
