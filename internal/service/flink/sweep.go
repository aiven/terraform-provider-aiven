//go:build sweep
// +build sweep

package flink

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_flink", &resource.Sweeper{
		Name:         "aiven_flink",
		F:            sweepFlinkServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepFlinkServices(region string) error {
	return sweep.SweepServices(region, "flink")
}
