//go:build sweep
// +build sweep

package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_opensearch", &resource.Sweeper{
		Name:         "aiven_opensearch",
		F:            sweepOpensearchServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepOpensearchServices(region string) error {
	return sweep.SweepServices(region, "opensearch")
}
