package kafka_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenKafka_import(t *testing.T) {
	project := os.Getenv("AIVEN_PROJECT_NAME")
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        fmt.Sprintf(kafkaImport, project, name),
				ResourceName:  "aiven_kafka.kafka",
				ImportState:   true,
				ImportStateId: "aiven-ci-kubernetes-operator/test-acc-lol",
			},
		},
	})
}

const kafkaImport = `
data "aiven_project" "project" {
  project = "%s"
}

resource "aiven_kafka" "kafka" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_connect_config {
      scheduled_rebalance_max_delay_ms = 10
    }
  }
}
`
