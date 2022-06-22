//go:build sweep
// +build sweep

package m3db

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_m3db", &resource.Sweeper{
		Name:         "aiven_m3db",
		F:            sweepM3dbServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepM3dbServices(region string) error {
	return sweep.SweepServices(region, "m3db")
}
