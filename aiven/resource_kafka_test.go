package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAiven_kafka(t *testing.T) {
	resourceName := "aiven_kafka.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_kafka.service"),
					testAccCheckAivenServiceKafkaAttributes("data.aiven_kafka.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_acl", "true"),
				),
			},
			{
				Config: testAccKafkaWithoutDefaultACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_kafka.service"),
					testAccCheckAivenServiceKafkaAttributes("data.aiven_kafka.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_acl", "false"),
				),
			},
		},
	})
}

func testAccKafkaResource(name string) string {
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
				kafka_version = "2.4"

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

		data "aiven_kafka" "service" {
			service_name = aiven_kafka.bar.service_name
			project = aiven_kafka.bar.project

			depends_on = [aiven_kafka.bar]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccKafkaWithoutDefaultACLResource(name string) string {
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
			default_acl = false
			kafka_user_config {
				kafka_rest = true
				kafka_connect = true
				schema_registry = true
				kafka_version = "2.4"
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
		data "aiven_kafka" "service" {
			service_name = aiven_kafka.bar.service_name
			project = aiven_kafka.bar.project

			depends_on = [aiven_kafka.bar]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
