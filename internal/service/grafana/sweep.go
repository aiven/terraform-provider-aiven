//go:build sweep
// +build sweep

package grafana

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_grafana", &resource.Sweeper{
		Name:         "aiven_grafana",
		F:            sweepGrafanaServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepGrafanaServices(region string) error {
	return sweep.SweepServices(region, "grafana")
}
