package staticip

import (
	"context"
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_static_ip", &resource.Sweeper{
		Name: "aiven_static_ip",
		F:    sweepStaticIPs(ctx),
		Dependencies: []string{
			"aiven_cassandra",
			"aiven_clickhouse",
			"aiven_flink",
			"aiven_grafana",
			"aiven_influxdb",
			"aiven_kafka",
			"aiven_kafka_connect",
			"aiven_kafka_mirrormaker",
			"aiven_m3db",
			"aiven_m3aggregator",
			"aiven_mysql",
			"aiven_opensearch",
			"aiven_pg",
			"aiven_redis",
		},
	})

}

func sweepStaticIPs(ctx context.Context) func(region string) error {
	return func(_ string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		r, err := client.StaticIPs.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of static_ips : %w", err)
		}

		for _, ip := range r.StaticIPs {
			err := client.StaticIPs.Delete(
				ctx,
				projectName,
				aiven.DeleteStaticIPRequest{
					StaticIPAddressID: ip.StaticIPAddressID,
				})
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting staticip: %w", err)
			}
		}

		return nil
	}
}
