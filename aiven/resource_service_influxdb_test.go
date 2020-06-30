package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"os"
	"testing"
)

// InfluxDB service tests
func TestAccAivenService_influxdb(t *testing.T) {
	t.Parallel()
	resourceName := "aiven_service.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfluxdbServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_service.service"),
					testAccCheckAivenServiceInfluxdbAttributes("data.aiven_service.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "influxdb"),
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

func testAccInfluxdbServiceResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "influxdb"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			influxdb_user_config {
				public_access {
					influxdb = true
				}
			}
		}
		
		data "aiven_service" "service" {
			service_name = aiven_service.bar.service_name
			project = aiven_project.foo.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}

func testAccCheckAivenServiceInfluxdbAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "influxdb" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["influxdb_user_config.0.public_access.0.influxdb"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.influxdb from Aiven")
		}

		return nil
	}
}
