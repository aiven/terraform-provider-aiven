//go:build sweep
// +build sweep

package clickhouse

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_clickhouse", &resource.Sweeper{
		Name:         "aiven_clickhouse",
		F:            sweepClickhouseServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepClickhouseServices(region string) error {
	return sweep.SweepServices(region, "clickhouse")
}
