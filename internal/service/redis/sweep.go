//go:build sweep
// +build sweep

package redis

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_redis", &resource.Sweeper{
		Name:         "aiven_redis",
		F:            sweepRedisServices,
		Dependencies: []string{"aiven_service_integration"},
	})
}

func sweepRedisServices(region string) error {
	return sweep.SweepServices(region, "redis")
}
