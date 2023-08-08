package kafka_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// MySQL service tests
func TestAccAivenService_mirrormaker(t *testing.T) {
	resourceName := "aiven_kafka_mirrormaker.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMirrorMakerServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceMirrorMakerAttributes("data.aiven_kafka_mirrormaker.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka_mirrormaker" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"

  kafka_mirrormaker_user_config {
    ip_filter = ["0.0.0.0/0"]

    kafka_mirrormaker {
      refresh_groups_interval_seconds = 600
      refresh_topics_enabled          = true
      refresh_topics_interval_seconds = 600
    }
  }
}

data "aiven_kafka_mirrormaker" "common" {
  service_name = aiven_kafka_mirrormaker.bar.service_name
  project      = aiven_kafka_mirrormaker.bar.project

  depends_on = [aiven_kafka_mirrormaker.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceMirrorMakerAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

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
