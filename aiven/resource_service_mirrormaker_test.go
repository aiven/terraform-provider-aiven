package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

// MySQL service tests
func TestAccAivenService_mirrormaker(t *testing.T) {
	t.Parallel()
	resourceName := "aiven_service.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMirrorMakerServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceMirrorMakerAttributes("data.aiven_service.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka_mirrormaker"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccMirrorMakerServiceResource(name string) string {
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

		data "aiven_service" "service" {
			service_name = aiven_service.bar.service_name
			project = aiven_project.foo.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}

func testAccCheckAivenServiceMirrorMakerAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "kafka_mirrormaker" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["kafka_mirrormaker_user_config.0.kafka_mirrormaker.0.refresh_groups_interval_seconds"] != "600" {
			return fmt.Errorf("expected to get a correct refresh_groups_interval_seconds from Aiven")
		}

		if a["kafka_mirrormaker_user_config.0.kafka_mirrormaker.0.refresh_topics_enabled"] != "true" {
			return fmt.Errorf("expected to get a correct refresh_topics_enabled from Aiven")
		}

		if a["kafka_mirrormaker_user_config.0.kafka_mirrormaker.0.refresh_topics_interval_seconds"] != "600" {
			return fmt.Errorf("expected to get a correct refresh_topics_interval_seconds from Aiven")
		}

		return nil
	}
}
