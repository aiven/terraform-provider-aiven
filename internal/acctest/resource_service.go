package acctest

import (
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_service", &resource.Sweeper{
		Name: "aiven_service",
		F:    sweepServices,
		Dependencies: []string{
			"aiven_database",
			"aiven_kafka_topic",
			"aiven_kafka_schema",
			"aiven_kafka_connector",
			"aiven_connection_pool",
			"aiven_service_integration",
		},
	})
}

func sweepServices(region string) error {
	client, err := SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name != os.Getenv("AIVEN_PROJECT_NAME") {
			continue
		}
		services, err := conn.Services.List(project.Name)
		if err != nil {
			return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
		}

		for _, s := range services {
			// if service termination_protection is on service cannot be deleted
			// update service and turn termination_protection off
			if s.TerminationProtection {
				_, err := conn.Services.Update(project.Name, s.Name, aiven.UpdateServiceRequest{
					Cloud:                 s.CloudName,
					MaintenanceWindow:     &s.MaintenanceWindow,
					Plan:                  s.Plan,
					ProjectVPCID:          s.ProjectVPCID,
					Powered:               true,
					TerminationProtection: false,
					UserConfig:            s.UserConfig,
				})

				if err != nil {
					return fmt.Errorf("error disabling `termination_protection` for service '%s' during sweep: %s", s.Name, err)
				}
			}

			if err := conn.Services.Delete(project.Name, s.Name); err != nil {
				if !aiven.IsNotFound(err) {
					return fmt.Errorf("error destroying service %s during sweep: %s", s.Name, err)
				}
			}
		}
	}
	return nil
}

func TestAccCheckAivenServiceCommonAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["cloud_name"] == "" {
			return fmt.Errorf("expected to get a cloud_name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service name from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["plan"] == "" {
			return fmt.Errorf("expected to get a plan from Aiven")
		}

		if a["service_uri"] == "" {
			return fmt.Errorf("expected to get a service_uri from Aiven")
		}

		if a["maintenance_window_dow"] != "monday" {
			return fmt.Errorf("expected to get a service.maintenance_window_dow from Aiven")
		}

		// Kafka service has no username and password
		if a["service_type"] != "kafka" {
			if a["service_password"] == "" {
				return fmt.Errorf("expected to get a service_password from Aiven")
			}

			if a["service_username"] == "" {
				return fmt.Errorf("expected to get a service_username from Aiven")
			}
		}

		if a["service_port"] == "" {
			return fmt.Errorf("expected to get a service_port from Aiven")
		}

		if a["service_host"] == "" {
			return fmt.Errorf("expected to get a service_host from Aiven")
		}

		if a["service_type"] == "" {
			return fmt.Errorf("expected to get a service_type from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["state"] != "RUNNING" {
			return fmt.Errorf("expected to get a correct state from Aiven")
		}

		if a["maintenance_window_time"] != "10:00:00" {
			return fmt.Errorf("expected to get a service.maintenance_window_time from Aiven")
		}

		if a["termination_protection"] == "" {
			return fmt.Errorf("expected to get a termination_protection from Aiven")
		}

		return nil
	}
}
