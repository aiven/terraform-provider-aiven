package deployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_flink_application_deployment", &resource.Sweeper{
		Name: "aiven_flink_application_deployment",
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

					deployments, err := client.ServiceFlinkListApplicationDeployments(ctx, projectName, s.ServiceName, app.Id)
					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving deployments for application %s: %w", app.Name, err)
					}

					for _, d := range deployments {
						if err := cancelAndDeleteDeployment(ctx, client, projectName, s.ServiceName, app.Id, d.Id); err != nil {
							return err
						}
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_flink"},
	})
}

func cancelAndDeleteDeployment(ctx context.Context, client avngen.Client, project, serviceName, applicationID, deploymentID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for {
		_, _ = client.ServiceFlinkCancelApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)

		if _, err := client.ServiceFlinkDeleteApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID); err == nil || avngen.IsNotFound(err) {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout deleting flink deployment %s: %w", deploymentID, ctx.Err())
		case <-time.After(time.Second):
		}
	}
}
