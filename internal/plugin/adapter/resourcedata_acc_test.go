package adapter_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestResourcePlanWithTimeouts: the timeouts block is not represented as a list of objects in the Plugin Framework.
// This test ensures that setting a timeout value for the create operation works.
func TestResourcePlanWithTimeouts(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.NoopProviderServer(),
		Steps: []resource.TestStep{{
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
			Config: `resource "aiven_project_vpc" "aiven_vpc" {
  project      = "my-project"
  cloud_name   = "aws-eu-central-1"
  network_cidr = "192.168.6.0/24"

  timeouts {
    create = "5m"
  }
}`,
		}},
	})
}
