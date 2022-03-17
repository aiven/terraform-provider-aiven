package kafka_test

import (
	"fmt"
	"os"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAiven_kafka_mirrormaker(t *testing.T) {
	resourceName := "aiven_kafka_mirrormaker.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMirrorMakerResource(rName),
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

func testAccMirrorMakerResource(name string) string {
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
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name)
}
