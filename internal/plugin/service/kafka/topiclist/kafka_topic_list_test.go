package topiclist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenKafkaTopicListDataSource(t *testing.T) {
	projectName := acc.ProjectName()
	kafkaName := acc.RandName("kafka")
	topicName1 := acc.RandName("topic")
	topicName2 := acc.RandName("topic")

	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		kafkaName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	topicsConfig := fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project      = %q
  service_name = %q
  topic_name   = %q
  partitions   = 1
  replication  = 2
}

resource "aiven_kafka_topic" "bar" {
  project      = %q
  service_name = %q
  topic_name   = %q
  partitions   = 3
  replication  = 2
}
`, projectName, kafkaName, topicName1, projectName, kafkaName, topicName2)

	dataSourceConfig := fmt.Sprintf(`
data "aiven_kafka_topic_list" "all" {
  project      = %q
  service_name = %q
  depends_on   = [aiven_kafka_topic.foo, aiven_kafka_topic.bar]
}
`, projectName, kafkaName)

	dataSourceName := "data.aiven_kafka_topic_list.all"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acc.TestAccPreCheck(t)
			t.Helper()
			if err := <-serviceIsReady; err != nil {
				t.Fatalf("failed to create test service: %s", err)
			}
		},
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create the topics before the datasource to ensure they are present when the datasource is read
				Config: topicsConfig,
			},
			{
				Config: topicsConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", projectName),
					resource.TestCheckResourceAttr(dataSourceName, "service_name", kafkaName),
					resource.TestCheckResourceAttrSet(dataSourceName, "topics.#"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "topics.*", map[string]string{
						"topic_name": topicName1,
						"state":      "ACTIVE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "topics.*", map[string]string{
						"topic_name": topicName2,
						"state":      "ACTIVE",
					}),
				),
			},
		},
	})
}
