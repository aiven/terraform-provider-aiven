package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

func TestAccAivenServiceIntegration_basic(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceIntegraitonResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-influxdb-%s", rName)),
				),
			},
			{
				Config:             testAccServiceIntegrationMirrorMakerResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServiceIntegrationResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar-pg" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-pg-%s"
			service_type = "pg"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				pg_version = 11

				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}

		resource "aiven_service" "bar-influxdb" {
				project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-influxdb-%s"
			service_type = "influxdb"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			influxdb_user_config {
				public_access {
					influxdb = true
				}
			}
		}

		resource "aiven_service_integration" "bar" {
			project = aiven_project.foo.project
			integration_type = "metrics"
			source_service_name = aiven_service.bar-pg.service_name
			destination_service_name = aiven_service.bar-influxdb.service_name
		}

		data "aiven_service_integration" "int" {
			project = aiven_service_integration.bar.project
			integration_type = aiven_service_integration.bar.integration_type
			source_service_name = aiven_service_integration.bar.source_service_name
			destination_service_name = aiven_service_integration.bar.destination_service_name
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name)
}

func testAccServiceIntegrationMirrorMakerResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "source" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "business-4"
			service_name = "test-acc-sr-source-%s"
			service_type = "kafka"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			kafka_user_config {
				kafka_version = "2.4"
				kafka {
				  group_max_session_timeout_ms = 70000
				  log_retention_bytes = 1000000000
				}
			}
		}
		
		resource "aiven_kafka_topic" "source" {
			project = aiven_project.foo.project
			service_name = aiven_service.source.service_name
			topic_name = "test-acc-topic-a-%s"
			partitions = 3
			replication = 2
		}

		resource "aiven_service" "target" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "business-4"
			service_name = "test-acc-sr-target-%s"
			service_type = "kafka"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			kafka_user_config {
				kafka_version = "2.4"
				kafka {
				  group_max_session_timeout_ms = 70000
				  log_retention_bytes = 1000000000
				}
			}
		}
		
		resource "aiven_kafka_topic" "target" {
			project = aiven_project.foo.project
			service_name = aiven_service.target.service_name
			topic_name = "test-acc-topic-b-%s"
			partitions = 3
			replication = 2
		}

		resource "aiven_service" "mm" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-mm-%s"
			service_type = "kafka_mirrormaker"
			
			kafka_mirrormaker_user_config {
				ip_filter = ["0.0.0.0/0"]

				kafka_mirrormaker {
					refresh_groups_interval_seconds = 600
					refresh_topics_enabled = true
					refresh_topics_interval_seconds = 600
				}
			}
		}

		resource "aiven_service_integration" "bar" {
			project = aiven_project.foo.project
			integration_type = "kafka_mirrormaker"
			source_service_name = aiven_service.source.service_name
			destination_service_name = aiven_service.mm.service_name
	
			kafka_mirrormaker_user_config {
				cluster_alias = "source"
			}
		}

		resource "aiven_service_integration" "i2" {
			project = aiven_project.foo.project
			integration_type = "kafka_mirrormaker"
			source_service_name = aiven_service.target.service_name
			destination_service_name = aiven_service.mm.service_name
	
			kafka_mirrormaker_user_config {
				cluster_alias = "target"
			}
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name, name, name, name)
}

func testAccCheckAivenServiceIntegraitonResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service_integration is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration" {
			continue
		}

		projectName, integrationID := splitResourceID2(rs.Primary.ID)
		i, err := c.ServiceIntegrations.Get(projectName, integrationID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if i != nil {
			return fmt.Errorf("service integration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAivenServiceIntegrationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["integration_type"] == "" {
			return fmt.Errorf("expected to get an integration_type from Aiven")
		}

		if a["source_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		if a["destination_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		return nil
	}
}
