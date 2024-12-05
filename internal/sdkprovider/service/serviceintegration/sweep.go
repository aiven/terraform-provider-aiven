package serviceintegration

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_service_integration", &resource.Sweeper{
		Name: "aiven_service_integration",
		F:    sweepServiceIntegrations(ctx),
	})

	sweep.AddTestSweepers("aiven_service_integration_endpoint", &resource.Sweeper{
		Name: "aiven_service_integration_endpoint",
		F:    sweepServiceIntegrationEndpoints(ctx),
	})
}

func sweepServiceIntegrations(ctx context.Context) func(region string) error {
	return func(_ string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		services, err := client.Services.List(ctx, projectName)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of service for a project `%s`: %w", projectName, err)
		}

		for _, service := range services {
			if len(service.Integrations) == 0 {
				continue
			}

			serviceIntegrations, err := client.ServiceIntegrations.List(ctx, projectName, service.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of service integration for service `%s`: %w", service.Name, err)
			}

			for _, serviceIntegration := range serviceIntegrations {
				err = client.ServiceIntegrations.Delete(ctx, projectName, serviceIntegration.ServiceIntegrationID)
				if common.IsCritical(err) {
					return fmt.Errorf(
						"unable to delete service integration `%s`: %w",
						serviceIntegration.ServiceIntegrationID,
						err,
					)
				}
			}
		}

		return nil
	}
}

func sweepServiceIntegrationEndpoints(ctx context.Context) func(region string) error {
	return func(_ string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		endpoints, err := client.ServiceIntegrationEndpoints.List(ctx, projectName)
		if err != nil {
			return err
		}

		for _, endpoint := range endpoints {
			err = client.ServiceIntegrationEndpoints.Delete(ctx, projectName, endpoint.EndpointID)
			if common.IsCritical(err) {
				return err
			}
		}

		return nil
	}
}
