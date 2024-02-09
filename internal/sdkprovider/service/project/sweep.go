package project

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_project", &resource.Sweeper{
		Name: "aiven_project",
		F:    sweepProjects(ctx),
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

	sweep.AddTestSweepers("aiven_billing_group", &resource.Sweeper{
		Name: "aiven_billing_group",
		F:    sweepBillingGroups(ctx),
		Dependencies: []string{
			"aiven_project",
		},
	})
}

func sweepProjects(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		projects, err := client.Projects.List(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of projects : %w", err)
		}

		for _, project := range projects {
			if strings.Contains(project.Name, "test-acc-") {
				if err := client.Projects.Delete(ctx, project.Name); err != nil {
					var e aiven.Error

					// project not found
					if errors.As(err, &e) && e.Status == 404 {
						continue
					}

					// project with open balance cannot be destroyed
					if strings.Contains(e.Message, "open balance") && e.Status == 403 {
						log.Printf("[DEBUG] project %s with open balance cannot be destroyed", project.Name)
						continue
					}

					return fmt.Errorf("error destroying project %s during sweep: %w", project.Name, err)
				}
			}
		}

		return nil
	}
}

func sweepBillingGroups(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		billingGroups, err := client.BillingGroup.ListAll(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of billing groups : %w", err)
		}

		for _, billingGroup := range billingGroups {
			if strings.Contains(billingGroup.BillingGroupName, "test-acc-") {
				if err := client.BillingGroup.Delete(ctx, billingGroup.Id); err != nil {
					// billing group not found
					var e aiven.Error
					if errors.As(err, &e) && e.Status == 404 {
						continue
					}

					return fmt.Errorf(
						"error destroying billing group %s during sweep: %w",
						billingGroup.BillingGroupName,
						err,
					)
				}
			}
		}

		return nil
	}
}
