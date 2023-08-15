//go:build sweep

package project

import (
	"fmt"
	"log"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aiven_project", &resource.Sweeper{
		Name: "aiven_project",
		F:    sweepProjects,
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

func sweepProjects(region string) error {
	client, err := sweep.SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if strings.Contains(project.Name, "test-acc-") {
			if err := conn.Projects.Delete(project.Name); err != nil {
				e := err.(aiven.Error)

				// project not found
				if e.Status == 404 {
					continue
				}

				// project with open balance cannot be destroyed
				if strings.Contains(e.Message, "open balance") && e.Status == 403 {
					log.Printf("[DEBUG] project %s with open balance cannot be destroyed", project.Name)
					continue
				}

				return fmt.Errorf("error destroying project %s during sweep: %s", project.Name, err)
			}
		}
	}

	return nil
}
