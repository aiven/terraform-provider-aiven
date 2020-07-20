package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"os"
	"testing"
)

func TestAccAiven_grafana(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_grafana.service"),
					testAccCheckAivenServiceGrafanaAttributes("data.aiven_grafana.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "grafana"),
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

func testAccGrafanaResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}
		
		resource "aiven_grafana" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-1"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			grafana_user_config {
				alerting_enabled = true
				
				public_access {
					grafana = true
				}
			}
		}
		
		data "aiven_grafana" "service" {
			service_name = aiven_grafana.bar.service_name
			project = aiven_project.foo.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}
