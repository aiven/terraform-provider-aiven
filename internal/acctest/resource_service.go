package acctest

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
