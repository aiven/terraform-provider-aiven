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
	t.Parallel()
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceIntegraitonResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-influxdb-%s", rName)),
				),
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
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name)
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
