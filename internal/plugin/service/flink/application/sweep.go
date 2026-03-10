package application

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

	sweep.AddTestSweepers("aiven_flink_application", &resource.Sweeper{
		Name: "aiven_flink_application",
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
				if s.ServiceType != "flink" {
					continue
				}

				if !strings.HasPrefix(s.ServiceName, sweep.DefaultPrefix) {
					continue
				}

				apps, err := client.ServiceFlinkListApplications(ctx, projectName, s.ServiceName)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving flink applications for service %s: %w", s.ServiceName, err)
				}

				for _, app := range apps {
					if !strings.HasPrefix(app.Name, sweep.DefaultPrefix) {
						continue
					}

					if _, err = client.ServiceFlinkDeleteApplication(ctx, projectName, s.ServiceName, app.Id); common.IsCritical(err) {
						return fmt.Errorf("error deleting flink application %s (%s): %w", app.Name, app.Id, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_flink_application_deployment"},
	})
}
