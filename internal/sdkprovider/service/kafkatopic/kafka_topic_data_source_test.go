package kafkatopic_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenDatasourceKafkaTopic_doesnt_exist this datasource shares Read() function with real "resource"
// This test makes sure the read func doesn't create missing topics as it does for "resources"
func TestAccAivenDatasourceKafkaTopic_doesnt_exist(t *testing.T) {
	project := acc.ProjectName()
	prefix := "test-tf-acc-" + acctest.RandString(7)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Kafka exists
				Config: testAccAivenDatasourceKafkaTopicDoesntExist(prefix, project, false),
				Check:  resource.TestCheckResourceAttr("aiven_kafka.kafka", "state", "RUNNING"),
			},
			{
				// Can't import unknown topic
				Config:      testAccAivenDatasourceKafkaTopicDoesntExist(prefix, project, true),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func testAccAivenDatasourceKafkaTopicDoesntExist(prefix, project string, addData bool) string {
	result := fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

resource "aiven_kafka" "kafka" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "%[1]s-kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
`, prefix, project)

	if addData {
		result += `
data "aiven_kafka_topic" "whatever" {
  project      = data.aiven_project.project.project
  service_name = aiven_kafka.kafka.service_name
  topic_name   = "whatever"
}`
	}
	return result
}
