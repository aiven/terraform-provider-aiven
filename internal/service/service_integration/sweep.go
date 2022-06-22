//go:build sweep
// +build sweep

package service_integration

import (
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_service_integration", &resource.Sweeper{
		Name: "aiven_service_integration",
		F:    sweepServiceIntegrations,
	})
}

func sweepServiceIntegrations(region string) error {
	client, err := sweep.SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	services, err := conn.Services.List(projectName)
	if err != nil && !aiven.IsNotFound(err) {
		return fmt.Errorf("error retrieving a list of service for a project `%s`: %s", projectName, err)
	}

	for _, service := range services {
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
