package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

// Cassandra service tests
func TestAccAivenService_cassandra(t *testing.T) {
	resourceName := "aiven_service.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCassandraServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_service.service"),
					testAccCheckAivenServiceCassandraAttributes("data.aiven_service.service"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "cassandra"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccCassandraServiceResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}
		
		resource "aiven_service" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "cassandra"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			cassandra_user_config {
				migrate_sstableloader = true		
	
				public_access {
					prometheus = true
				}
			}
		}
		
		data "aiven_service" "service" {
			service_name = aiven_service.bar.service_name
			project = aiven_service.bar.project

			depends_on = [aiven_service.bar]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceCassandraAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "cassandra" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["cassandra_user_config.0.public_access.0.prometheus"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["cassandra_user_config.0.service_to_fork_from"] != "" {
			return fmt.Errorf("expected to get a correct public_access.service_to_fork_from from Aiven")
		}

		if a["cassandra_user_config.0.migrate_sstableloader"] != "true" {
			return fmt.Errorf("expected to get a correct migrate_sstableloader from Aiven")
		}

		return nil
	}
}
