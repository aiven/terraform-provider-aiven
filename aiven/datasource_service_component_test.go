package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenServiceComponentDataSource_basic(t *testing.T) {
	datasourceKafka := "data.aiven_service_component.kafka"
	datasourceKafkaConnect := "data.aiven_service_component.kafka_connect"
	datasourceKafkaRest := "data.aiven_service_component.kafka_rest"
	datasourceKafkaRegistry := "data.aiven_service_component.schema_registry"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComponentDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceKafka, "project"),
					testAccServiceComponentAttributes(datasourceKafka, "kafka", "dynamic"),
					testAccServiceComponentKafkaAuthenticationMethod(datasourceKafka),
					// Kafka Connect
					testAccServiceComponentAttributes(datasourceKafkaConnect, "kafka_connect", "public"),
					// Kafka Rest
					testAccServiceComponentAttributes(datasourceKafkaRest, "kafka_rest", "public"),
					// Kafka Registry
					testAccServiceComponentAttributes(datasourceKafkaRegistry, "schema_registry", "dynamic"),
				),
			},
		},
	})
}

func testAccServiceComponentAttributes(n, component, route string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] != os.Getenv("AIVEN_PROJECT_NAME") {
			return fmt.Errorf("expected to get a corect project name from Aiven got: " + a["project"])
		}

		if a["component"] != component {
			return fmt.Errorf("expected to get a corect component from Aiven")
		}

		if a["route"] != route {
			return fmt.Errorf("expected to get a corect route from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["host"] == "" {
			return fmt.Errorf("expected to get a host from Aiven")
		}

		if a["usage"] == "" {
			return fmt.Errorf("expected to get a usage from Aiven")
		}

		return nil
	}
}

func testAccServiceComponentKafkaAuthenticationMethod(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["kafka_authentication_method"] == "" {
			return fmt.Errorf("expected to get a kafka_authentication_method from Aiven")
		}

		return nil
	}
}

func testAccServiceComponentDataSource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}

		resource "aiven_kafka" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "business-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"

			kafka_user_config {
				kafka_rest = true
				kafka_connect = true
				schema_registry = true

				kafka {
					group_max_session_timeout_ms = 70000
					log_retention_bytes = 1000000000
				}

				public_access {
					kafka_rest = true
					kafka_connect = true
				}
			}
		}

		data "aiven_service_component" "kafka" {
			project = aiven_kafka.bar.project
			service_name = aiven_kafka.bar.service_name
			component = "kafka"
			route = "dynamic"
			kafka_authentication_method = "certificate"
			
			depends_on = [ aiven_kafka.bar ]
		}
		
		data "aiven_service_component" "kafka_connect" {
			project = aiven_kafka.bar.project
			service_name = aiven_kafka.bar.service_name
			component = "kafka_connect"
			route = "public"

			depends_on = [ aiven_kafka.bar ]
		}

		data "aiven_service_component" "kafka_rest" {
			project = aiven_kafka.bar.project
			service_name = aiven_kafka.bar.service_name
			component = "kafka_rest"
			route = "public"

			depends_on = [ aiven_kafka.bar ]
		}
		
		data "aiven_service_component" "schema_registry" {
			project = aiven_kafka.bar.project
			service_name = aiven_kafka.bar.service_name
			component = "schema_registry"
			route = "dynamic"

			depends_on = [ aiven_kafka.bar ]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
