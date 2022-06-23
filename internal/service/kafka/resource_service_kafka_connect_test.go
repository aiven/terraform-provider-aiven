package kafka_test

import (
	"fmt"
	"os"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Kafka Connect service tests
func TestAccAivenService_kafkaconnect(t *testing.T) {
	resourceName := "aiven_kafka_connect.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaConnectServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_kafka_connect.common"),
					testAccCheckAivenServiceKafkaConnectAttributes("data.aiven_kafka_connect.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka_connect"),
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

func testAccKafkaConnectServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka_connect" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = true
    }
  }
}

data "aiven_kafka_connect" "common" {
  service_name = aiven_kafka_connect.bar.service_name
  project      = aiven_kafka_connect.bar.project

  depends_on = [aiven_kafka_connect.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceKafkaConnectAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["kafka_connect_user_config.0.kafka_connect.0.consumer_isolation_level"] != "read_committed" {
			return fmt.Errorf("expected to get a correct consumer_isolation_level from Aiven")
		}

		if a["kafka_connect_user_config.0.kafka_connect.0.consumer_max_poll_records"] != "" {
			return fmt.Errorf("expected to get a correct consumer_max_poll_records from Aiven")
		}

		if a["kafka_connect_user_config.0.kafka_connect.0.offset_flush_interval_ms"] != "" {
			return fmt.Errorf("expected to get a correct offset_flush_interval_ms from Aiven")
		}

		if a["kafka_connect_user_config.0.public_access.0.kafka_connect"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.kafka_connect from Aiven")
		}

		if a["kafka_connect_user_config.0.public_access.0.prometheus"] != "" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		return nil
	}
}
