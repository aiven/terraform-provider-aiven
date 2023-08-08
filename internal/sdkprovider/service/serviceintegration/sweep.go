//go:build sweep

package serviceintegration

import (
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aiven_service_integration", &resource.Sweeper{
		Name: "aiven_service_integration",
		F:    sweepServiceIntegrations,
	})

	resource.AddTestSweepers("aiven_service_integration_endpoint", &resource.Sweeper{
		Name: "aiven_service_integration_endpoint",
		F:    sweepServiceIntegrationEndpoints,
	})
}

func sweepServiceIntegrations(region string) error {
	client, err := sweep.SharedClient(region)
	if err != nil {
		return err
	}

	conn := client.(*aiven.Client)

	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	services, err := conn.Services.List(projectName)
	if err != nil && !aiven.IsNotFound(err) {
		return fmt.Errorf("error retrieving a list of service for a project `%s`: %s", projectName, err)
	}

	for _, service := range services {
		if len(service.Integrations) == 0 {
			continue
		}

		serviceIntegrations, err := conn.ServiceIntegrations.List(projectName, service.Name)
		if err != nil {
			return fmt.Errorf("error retrieving a list of service integration for service `%s`: %s", service.Name, err)
		}
		for _, serviceIntegration := range serviceIntegrations {
			if err := conn.ServiceIntegrations.Delete(projectName, serviceIntegration.ServiceIntegrationID); err != nil {
				if !aiven.IsNotFound(err) {
					return fmt.Errorf("unable to delete service integration `%s`: %s", serviceIntegration.ServiceIntegrationID, err)
				}
			}
		}
	}

	return nil
}

func sweepServiceIntegrationEndpoints(region string) error {
	client, err := sweep.SharedClient(region)
	if err != nil {
		return err
	}

	conn := client.(*aiven.Client)

	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	endpoints, err := conn.ServiceIntegrationEndpoints.List(projectName)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		err = conn.ServiceIntegrationEndpoints.Delete(projectName, endpoint.EndpointID)
		if err != nil && !aiven.IsNotFound(err) {
			return err
		}
	}

	return nil
}
