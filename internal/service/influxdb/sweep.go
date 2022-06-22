//go:build sweep
// +build sweep

package influxdb

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_influxdb", &resource.Sweeper{
		Name:         "aiven_influxdb",
		F:            sweepInfluxdbServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepInfluxdbServices(region string) error {
	return sweep.SweepServices(region, "influxdb")
}
