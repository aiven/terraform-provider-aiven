// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_kafka_topic", &resource.Sweeper{
		Name: "aiven_kafka_topic",
		F:    sweepKafkaTopics,
	})
}

func sweepKafkaTopics(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name == os.Getenv("AIVEN_PROJECT_NAME") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				if service.Type != "kafka" {
					continue
				}

				topics, err := conn.KafkaTopics.List(project.Name, service.Name)
				if err != nil {
					log.Printf("[ERROR] error retrieving a list of kafka topics for a service `%s`: %s", service.Name, err)
					continue
				}

				for _, topic := range topics {
					err = conn.KafkaTopics.Delete(project.Name, service.Name, topic.TopicName)
					if err != nil {
						return fmt.Errorf("error destroying kafka topic %s during sweep: %s", topic.TopicName, err)
					}
				}
			}
		}
	}

	return nil
}

func TestAccAivenKafkaTopic(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_kafka_topic.foo"

	t.Run("kafka topic basic", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaTopicResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaTopicAttributes("data.aiven_kafka_topic.topic"),
						resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "topic_name", fmt.Sprintf("test-acc-topic-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					),
				},
			},
		})
	})

	t.Run("kafka topic custom timeouts", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaTopicCustomTimeoutsResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaTopicAttributes("data.aiven_kafka_topic.topic"),
						resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "topic_name", fmt.Sprintf("test-acc-topic-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
						resource.TestCheckResourceAttr(resourceName, "retention_hours", "100"),
					),
				},
			},
		})
	})

	t.Run("kafka topic termination protection", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config:                    testAccKafkaTopicTerminationProtectionResource(rName),
					PreventPostDestroyRefresh: true,
					ExpectNonEmptyPlan:        true,
					PlanOnly:                  true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "topic_name", fmt.Sprintf("test-acc-topic-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
						resource.TestCheckNoResourceAttr(resourceName, "retention_hours"),
					),
				},
			},
		})
	})

	t.Run("kafka topic many topics", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafka451TopicResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaTopicAttributes("data.aiven_kafka_topic.topic"),
						resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "topic_name", fmt.Sprintf("test-acc-topic-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
					),
				},
			},
		})
	})
}

func testAccKafka451TopicResource(name string) string {
	return testAccKafkaTopicResource(name) + `
		resource "aiven_kafka_topic" "more" {
		  count = 450
		
		  project      = data.aiven_project.foo.project
		  service_name = aiven_kafka.bar.service_name
		  topic_name   = "test-acc-topic-${count.index}"
		  partitions   = 3
		  replication  = 2
		}`

}

func testAccKafkaTopicResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_kafka" "bar" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		}
		
		resource "aiven_kafka_topic" "foo" {
		  project                  = data.aiven_project.foo.project
		  service_name             = aiven_kafka.bar.service_name
		  topic_name               = "test-acc-topic-%s"
		  partitions               = 3
		  replication              = 2
		  retention_hours          = 1
		  retention_bytes          = -1
		  minimum_in_sync_replicas = 2
		
		  config {
		    flush_ms                       = 10
		    unclean_leader_election_enable = true
		    cleanup_policy                 = "compact"
		    min_cleanable_dirty_ratio      = 0.01
		    delete_retention_ms            = 50000
		  }
		}
		
		data "aiven_kafka_topic" "topic" {
		  project      = aiven_kafka_topic.foo.project
		  service_name = aiven_kafka_topic.foo.service_name
		  topic_name   = aiven_kafka_topic.foo.topic_name
		
		  depends_on = [aiven_kafka_topic.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccKafkaTopicCustomTimeoutsResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_kafka" "bar" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  timeouts {
		    create = "25m"
		    update = "20m"
		  }
		}
		
		resource "aiven_kafka_topic" "foo" {
		  project         = data.aiven_project.foo.project
		  service_name    = aiven_kafka.bar.service_name
		  topic_name      = "test-acc-topic-%s"
		  partitions      = 3
		  replication     = 2
		  retention_hours = 100
		
		  timeouts {
		    create = "15m"
		    read   = "15m"
		  }
		}
		
		data "aiven_kafka_topic" "topic" {
		  project      = aiven_kafka_topic.foo.project
		  service_name = aiven_kafka_topic.foo.service_name
		  topic_name   = aiven_kafka_topic.foo.topic_name
		
		  depends_on = [aiven_kafka_topic.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccKafkaTopicTerminationProtectionResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_kafka" "bar" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  kafka_user_config {
		    kafka {
		      group_max_session_timeout_ms = 70000
		      log_retention_bytes          = 1000000000
		    }
		  }
		}
		
		resource "aiven_kafka_topic" "foo" {
		  project                = data.aiven_project.foo.project
		  service_name           = aiven_kafka.bar.service_name
		  topic_name             = "test-acc-topic-%s"
		  partitions             = 3
		  replication            = 2
		  termination_protection = true
		}
		
		data "aiven_kafka_topic" "topic" {
		  project      = aiven_kafka_topic.foo.project
		  service_name = aiven_kafka_topic.foo.service_name
		  topic_name   = aiven_kafka_topic.foo.topic_name
		
		  depends_on = [aiven_kafka_topic.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenKafkaTopicAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] kafka topic attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["topic_name"] == "" {
			return fmt.Errorf("expected to get a topic_name from Aiven")
		}

		if a["partitions"] == "" {
			return fmt.Errorf("expected to get partitions from Aiven")
		}

		if a["replication"] == "" {
			return fmt.Errorf("expected to get a replication from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenKafkaTopicResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each kafka topic is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_topic" {
			continue
		}

		project, serviceName, topicName := schemautil.SplitResourceID3(rs.Primary.ID)

		_, err := c.Services.Get(project, serviceName)
		if err != nil {
			if aiven.IsNotFound(err) {
				return nil
			}
			return err
		}

		t, err := c.KafkaTopics.Get(project, serviceName, topicName)
		if err != nil {
			if aiven.IsNotFound(err) {
				return nil
			}
			return err
		}

		if t != nil {
			return fmt.Errorf("kafka topic (%s) still exists, id %s", topicName, rs.Primary.ID)
		}
	}

	return nil
}

func TestPartitions(t *testing.T) {
	type args struct {
		numPartitions int
	}
	tests := []struct {
		name           string
		args           args
		wantPartitions []*aiven.Partition
	}{
		{
			"basic",
			args{numPartitions: 3},
			[]*aiven.Partition{{}, {}, {}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPartitions := partitions(tt.args.numPartitions); !reflect.DeepEqual(gotPartitions, tt.wantPartitions) {
				t.Errorf("partitions() = %v, want %v", gotPartitions, tt.wantPartitions)
			}
		})
	}
}
