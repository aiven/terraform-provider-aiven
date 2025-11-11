//go:build backwardcompat

package kafkatopic_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenKafkaTopicBackwardCompatibility verifies that the Kafka topic resource
// maintains backward compatibility with previous provider versions.
//
// This test ensures that resources created with the last stable published provider version
// can be managed by the current development version without any changes.
func TestAccAivenKafkaTopicBackwardCompatibility(t *testing.T) {
	var (
		projectName = acc.ProjectName()
		rName       = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-kafka-bc-%s", rName)
		topicName   = fmt.Sprintf("test-topic-bc-%s", rName)
	)

	config := fmt.Sprintf(`
resource "aiven_kafka" "kafka" {
  project                 = %[1]q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_kafka_topic" "test" {
  project                = aiven_kafka.kafka.project
  service_name           = aiven_kafka.kafka.service_name
  topic_name             = %[3]q
  partitions             = 3
  replication            = 2
  topic_description      = "Test topic for backward compatibility"
  termination_protection = false

  tag {
    key   = "environment"
    value = "test"
  }

  tag {
    key   = "purpose"
    value = "backward-compat-testing"
  }

  config {
    cleanup_policy                 = "delete"
    compression_type               = "producer"
    retention_ms                   = "604800000"
    retention_bytes                = "1073741824"
    min_insync_replicas            = "1"
    max_message_bytes              = "1048576"
    message_timestamp_type         = "CreateTime"
    unclean_leader_election_enable = false
    preallocate                    = false
    segment_bytes                  = "536870912"
  }

  depends_on = [aiven_kafka.kafka]
}`, projectName, serviceName, topicName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig: config,
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("aiven_kafka.kafka", "state", "RUNNING"),
				resource.TestCheckResourceAttrSet("aiven_kafka.kafka", "id"),

				resource.TestCheckResourceAttrSet("aiven_kafka_topic.test", "id"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "topic_name", topicName),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "partitions", "3"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "replication", "2"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "topic_description", "Test topic for backward compatibility"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "termination_protection", "false"),

				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "tag.#", "2"),

				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.cleanup_policy", "delete"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.compression_type", "producer"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.retention_ms", "604800000"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.retention_bytes", "1073741824"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.min_insync_replicas", "1"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.max_message_bytes", "1048576"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.message_timestamp_type", "CreateTime"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.unclean_leader_election_enable", "false"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.preallocate", "false"),
				resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.segment_bytes", "536870912"),

				func(s *terraform.State) error {
					// Wait for topic to be fully available in the API after creation.
					// Without this sleep, the API may not immediately reflect the topic's
					// final state, causing the plan check in the next step to detect drift
					// (non-empty plan) when switching to the new provider version.
					// This is an API eventual consistency issue, not a provider bug.
					time.Sleep(10 * time.Second)

					return nil
				},
			),
		}),
	})
}
