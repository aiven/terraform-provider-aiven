//go:build sweep
// +build sweep

package pg

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_pg", &resource.Sweeper{
		Name:         "aiven_pg",
		F:            sweepPGServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip", "aiven_static_ip"},
	})
}

func sweepPGServices(region string) error {
	return sweep.SweepServices(region, "pg")
}
