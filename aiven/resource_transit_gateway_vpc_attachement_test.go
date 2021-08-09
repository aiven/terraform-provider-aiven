package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenTransitGatewayVPCAttachment_basic(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" ||
		os.Getenv("AWS_TRANSIT_GATEWAY_ID") == "" ||
		os.Getenv("AWS_ACCOUNT_ID") == "" {
		t.Skip("env variables AWS_REGION, AWS_TRANSIT_GATEWAY_ID and AWS_ACCOUNT_ID required to run this test")
	}

	resourceName := "aiven_transit_gateway_vpc_attachment.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentResource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentAttributes("data.aiven_transit_gateway_vpc_attachment.att"),
					resource.TestCheckResourceAttr(resourceName, "peer_cloud_account", os.Getenv("AWS_ACCOUNT_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_vpc", os.Getenv("AWS_TRANSIT_GATEWAY_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_region", os.Getenv("AWS_REGION")),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentResource() string {
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

		resource "aiven_transit_gateway_vpc_attachment" "foo" {
			vpc_id = aiven_project_vpc.bar.id
			peer_cloud_account = "%s"
			peer_vpc = "%s"
			peer_region = "%s"
			user_peer_network_cidrs = [ "172.31.0.0/16" ]

			timeouts {
				create = "10m"
			}
		}

		data "aiven_transit_gateway_vpc_attachment" "att" {
			vpc_id = aiven_transit_gateway_vpc_attachment.foo.vpc_id
			peer_cloud_account = aiven_transit_gateway_vpc_attachment.foo.peer_cloud_account
			peer_vpc = aiven_transit_gateway_vpc_attachment.foo.peer_vpc

			depends_on = [aiven_transit_gateway_vpc_attachment.foo]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ACCOUNT_ID"),
		os.Getenv("AWS_TRANSIT_GATEWAY_ID"),
		os.Getenv("AWS_REGION"))
}

func testAccCheckTransitGatewayVPCAttachmentAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["vpc_id"] == "" {
			return fmt.Errorf("expected to get a vpc_id name from Aiven")
		}

		if a["peer_cloud_account"] == "" {
			return fmt.Errorf("expected to get a peer_cloud_account from Aiven")
		}

		if a["peer_vpc"] == "" {
			return fmt.Errorf("expected to get a peer_vpc from Aiven")
		}

		if a["peer_region"] == "" {
			return fmt.Errorf("expected to get a peer_region from Aiven")
		}

		if a["user_peer_network_cidrs.0"] == "" {
			return fmt.Errorf("expected to get a user_peer_network_cidrs from Aiven")
		}

		if a["state"] == "" {
			return fmt.Errorf("expected to get a state from Aiven")
		}

		return nil
	}
}
