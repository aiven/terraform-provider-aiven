package acctest_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// ExampleBackwardCompatibilitySteps demonstrates how to create a backward compatibility test.
func ExampleBackwardCompatibilitySteps() {
	t := &testing.T{}

	// define your Terraform configuration
	config := `
resource "aiven_kafka_topic" "test" {
  project      = "my-project"
  service_name = "my-kafka"
  topic_name   = "test-topic"
  partitions   = 3
  replication  = 2

  tag {
    key   = "environment"
    value = "production"
  }

  config {
    retention_ms        = "604800000"
    min_insync_replicas = "2"
  }
}`

	// create test steps using the helper
	steps := acctest.BackwardCompatibilitySteps(t, acctest.BackwardCompatConfig{
		TFConfig: config,
		// Optional: test against specific version instead of latest
		// OldProviderVersion: "4.25.0",
		Checks: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("aiven_kafka_topic.test", "partitions", "3"),
			resource.TestCheckResourceAttr("aiven_kafka_topic.test", "replication", "2"),
			resource.TestCheckResourceAttr("aiven_kafka_topic.test", "tag.#", "1"),
			resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.retention_ms", "604800000"),
			resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.min_insync_replicas", "2"),
		),
	})

	// use these steps in your resource.TestCase
	fmt.Printf("Created %d test steps for backward compatibility testing\n", len(steps))
	// Output: Created 2 test steps for backward compatibility testing
}
