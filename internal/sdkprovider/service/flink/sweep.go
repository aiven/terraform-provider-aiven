package flink

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

	sweep.AddServiceSweeper("flink")

	sweep.AddTestSweepers("aiven_flink_jar_application", &resource.Sweeper{
		Name: "aiven_flink_jar_application",
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

				apps, err := client.ServiceFlinkListJarApplications(ctx, projectName, s.ServiceName)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving flink jar applications for service %s: %w", s.ServiceName, err)
				}

				for _, app := range apps {
					if !strings.HasPrefix(app.Name, sweep.DefaultPrefix) {
						continue
					}

					if _, err = client.ServiceFlinkDeleteJarApplication(ctx, projectName, s.ServiceName, app.Id); common.IsCritical(err) {
						return fmt.Errorf("error deleting flink jar application %s (%s): %w", app.Name, app.Id, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_flink_jar_application_deployment"},
	})

	sweep.AddTestSweepers("aiven_flink_jar_application_deployment", &resource.Sweeper{
		Name: "aiven_flink_jar_application_deployment",
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

				apps, err := client.ServiceFlinkListJarApplications(ctx, projectName, s.ServiceName)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving flink jar applications for service %s: %w", s.ServiceName, err)
				}

				for _, app := range apps {
					if !strings.HasPrefix(app.Name, sweep.DefaultPrefix) {
						continue
					}

					deployments, err := client.ServiceFlinkListJarApplicationDeployments(ctx, projectName, s.ServiceName, app.Id)
					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving deployments for jar application %s: %w", app.Name, err)
					}

					for _, d := range deployments {
						if err := cancelAndDeleteJarDeployment(ctx, client, projectName, s.ServiceName, app.Id, d.Id); err != nil {
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

func cancelAndDeleteJarDeployment(ctx context.Context, client avngen.Client, project, serviceName, applicationID, deploymentID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for {
		_, _ = client.ServiceFlinkCancelJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)

		if _, err := client.ServiceFlinkDeleteJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID); err == nil || avngen.IsNotFound(err) {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout deleting flink jar deployment %s: %w", deploymentID, ctx.Err())
		case <-time.After(time.Second):
		}
	}
}
