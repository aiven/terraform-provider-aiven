package transitgatewayvpcattachment_test

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	transitGatewayAttachmentResourceName   = "aiven_transit_gateway_vpc_attachment.attachment"
	transitGatewayAttachmentDataSourceName = "data.aiven_transit_gateway_vpc_attachment.attachment"
)

var liveAttachmentStatePattern = regexp.MustCompile(`^(ACTIVE|PENDING_PEER)$`)

func testAccCheckAivenTransitGatewayVPCAttachmentIDParts(want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attachment, ok := s.RootModule().Resources[transitGatewayAttachmentResourceName]
		if !ok {
			return fmt.Errorf("transit gateway VPC attachment not found in state")
		}

		if got := len(strings.Split(attachment.Primary.ID, "/")); got != want {
			return fmt.Errorf("expected transit gateway VPC attachment ID with %d parts, got %d: %q", want, got, attachment.Primary.ID)
		}
		return nil
	}
}
