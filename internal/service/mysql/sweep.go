//go:build sweep
// +build sweep

package mysql

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_mysql", &resource.Sweeper{
		Name:         "aiven_mysql",
		F:            sweepMysqlServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepMysqlServices(region string) error {
	return sweep.SweepServices(region, "mysql")
}
