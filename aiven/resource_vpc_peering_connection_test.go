package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccAivenVPCPeeringConnection_basic(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" ||
		os.Getenv("AWS_VPC_ID") == "" ||
		os.Getenv("AWS_ACCOUNT_ID") == "" {
		t.Skip("env variables AWS_REGION, AWS_VPC_ID and AWS_ACCOUNT_ID required to run this test")
	}

	resourceName := "aiven_vpc_peering_connection.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
			project = data.aiven_project.foo.project
			cloud_name = "aws-%s"
			network_cidr = "10.0.0.0/24"

			timeouts {
				create = "5m"
			}
		}

		resource "aiven_vpc_peering_connection" "foo" {
			vpc_id = aiven_project_vpc.bar.id
			peer_cloud_account = "%s"
			peer_vpc = "%s"
			peer_region = "%s"

			timeouts {
				create = "10m"
			}
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ACCOUNT_ID"),
		os.Getenv("AWS_VPC_ID"),
		os.Getenv("AWS_REGION"))
}
