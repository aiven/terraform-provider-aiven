//go:build sweep
// +build sweep

package cassandra

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_cassandra", &resource.Sweeper{
		Name:         "aiven_cassandra",
		F:            sweepCassandraServices,
		Dependencies: []string{"aiven_service_integration"},
	})
}

func sweepCassandraServices(region string) error {
	return sweep.SweepServices(region, "cassandra")
}
