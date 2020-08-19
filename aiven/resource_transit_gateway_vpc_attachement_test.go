package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccAivenTransitGatewayVPCAttachment_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             testAccTransitGatewayVPCAttachmentResource(rName),
				Check:              resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_vpc" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			network_cidr = "192.168.0.0/24"
		}

		resource "aiven_transit_gateway_vpc_attachment" "foo" {
			vpc_id = aiven_project_vpc.bar.id
			peer_cloud_account = "<PEER_ACCOUNT_ID>"
			peer_vpc = "google-project1"
			peer_region = "google-europe-west1"
			user_peer_network_cidrs = [ "10.0.0.0/24" ]
		}
		`, name)
}
