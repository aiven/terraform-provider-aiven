//go:build sweep
// +build sweep

package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_kafka", &resource.Sweeper{
		Name:         "aiven_kafka",
		F:            sweepKafkaServices,
		Dependencies: []string{"aiven_service_integration", "aiven_static_ip"},
	})
}

func sweepKafkaServices(region string) error {
	return sweep.SweepServices(region, "kafka")
}
