package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddServiceSweeper("kafka")
	sweep.AddServiceSweeper("kafka_mirrormaker")
	sweep.AddServiceSweeper("kafka_connect")
	sweep.AddServiceSweeper("kafka_connector")

	sweep.AddTestSweepers("aiven_mirrormaker_replication_flow", &resource.Sweeper{
		Name: "aiven_mirrormaker_replication_flow",
		F: func(_ string) error {
			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}

			projectName := sweep.ProjectName()

			services, err := client.ServiceList(ctx, projectName)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving services for project %s: %w", projectName, err)
			}

			for _, s := range services {
				if s.ServiceType != "kafka_mirrormaker" {
					continue
				}

				if !strings.HasPrefix(s.ServiceName, sweep.DefaultPrefix) {
					continue
				}

				flows, err := client.ServiceKafkaMirrorMakerGetReplicationFlows(ctx, projectName, s.ServiceName)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving replication flows for service %s: %w", s.ServiceName, err)
				}

				for _, flow := range flows {
					if err = client.ServiceKafkaMirrorMakerDeleteReplicationFlow(ctx, projectName, s.ServiceName, flow.SourceCluster, flow.TargetCluster); common.IsCritical(err) {
						return fmt.Errorf("error deleting replication flow %s->%s in service %s: %w", flow.SourceCluster, flow.TargetCluster, s.ServiceName, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_kafka_mirrormaker"},
	})
}
