package flink_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAiven_flink(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	serviceName := fmt.Sprintf("test-acc-flink-%s", randString())
	manifest := fmt.Sprintf(`
		variable "project_name" {
		  type    = string
		  default = "%s"
		}
		variable "service_name" {
		  type    = string
		  default = "%s"
		}
		
		resource "aiven_flink" "bar" {
		  project                 = var.project_name
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = var.service_name
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  tag {
		    key   = "test"
		    value = "val"
		  }
		
		  flink_user_config {
		    number_of_task_slots = 10
		    parallelism_default  = 2
		    restart_strategy     = "failure-rate"
		  }
		}
		
		data "aiven_flink" "service" {
		  service_name = aiven_flink.bar.service_name
		  project      = aiven_flink.bar.project
		}`,

		projectName,
		serviceName,
	)

	manifestDoubleTags := fmt.Sprintf(`
		variable "project_name" {
		  type    = string
		  default = "%s"
		}
		variable "service_name" {
		  type    = string
		  default = "%s"
		}
		
		resource "aiven_flink" "bar" {
		  project                 = var.project_name
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = var.service_name
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  tag {
		    key   = "test"
		    value = "val"
		  }
		  tag {
		    key   = "test"
		    value = "val2"
		  }
		
		  flink_user_config {
		    number_of_task_slots = 10
		    parallelism_default  = 2
		    restart_strategy     = "failure-rate"
		  }
		}
		
		data "aiven_flink" "service" {
		  service_name = aiven_flink.bar.service_name
		  project      = aiven_flink.bar.project
		}`,

		projectName,
		serviceName,
	)
	resourceName := "aiven_flink.bar"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenFlinkJobsAndTableResourcesDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_flink.service"),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "flink_user_config.0.number_of_task_slots", "10"),
					resource.TestCheckResourceAttr(resourceName, "flink_user_config.0.parallelism_default", "2"),
					resource.TestCheckResourceAttr(resourceName, "flink_user_config.0.restart_strategy", "failure-rate"),
				),
			},
			{
				Config:             manifestDoubleTags,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func TestAccAiven_flink_kafka_to_pg(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	flinkServiceName := fmt.Sprintf("test-acc-flink-%s", randString())
	kafkaServiceName := fmt.Sprintf("test-acc-flink-kafka-%s", randString())
	postgresServiceName := fmt.Sprintf("test-acc-flink-postgres-%s", randString())
	sourceTopicName := fmt.Sprintf("test-acc-flink-kafka-source-topic-%s", randString())
	sourceTableName := fmt.Sprintf("test_acc_flink_kafka_source_table_%s", randString())
	sinkTableName := fmt.Sprintf("test_acc_flink_kafka_sink_table_%s", randString())
	sinkJdbcTableName := fmt.Sprintf("test_acc_flink_kafka_source_jdbc_table_%s", randString())
	jobName := fmt.Sprintf("test_acc_flink_job_%s", randString())
	manifest := fmt.Sprintf(`
		variable "project_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_flink" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_kafka" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_pg" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_topic_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_jdbc_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "job_name" {
		  type    = string
		  default = "%s"
		}
		
		resource "aiven_flink" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-4"
		  service_name = var.service_name_flink
		}
		
		resource "aiven_kafka" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-2"
		  service_name = var.service_name_kafka
		}
		
		resource "aiven_pg" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-4"
		  service_name = var.service_name_pg
		}
		
		resource "aiven_kafka_topic" "source" {
		  project      = aiven_kafka.testing.project
		  service_name = aiven_kafka.testing.service_name
		  topic_name   = var.source_topic_name
		  replication  = 2
		  partitions   = 2
		}
		
		resource "aiven_service_integration" "flinkkafka" {
		  project                  = aiven_flink.testing.project
		  integration_type         = "flink"
		  destination_service_name = aiven_flink.testing.service_name
		  source_service_name      = aiven_kafka.testing.service_name
		}
		
		resource "aiven_service_integration" "flinkpg" {
		  project                  = aiven_flink.testing.project
		  integration_type         = "flink"
		  destination_service_name = aiven_flink.testing.service_name
		  source_service_name      = aiven_pg.testing.service_name
		}
		
		resource "aiven_flink_table" "source" {
		  project        = aiven_flink.testing.project
		  service_name   = aiven_flink.testing.service_name
		  integration_id = aiven_service_integration.flinkkafka.integration_id
		  table_name     = var.source_table_name
		  kafka_topic    = aiven_kafka_topic.source.topic_name
		  schema_sql     = <<EOF
		    cpu INT,
		    node INT,
		    occurred_at TIMESTAMP(3) METADATA FROM 'timestamp',
		    WATERMARK FOR occurred_at AS occurred_at - INTERVAL '5' SECOND
		  EOF
		}
		
		resource "aiven_flink_table" "sink" {
		  project        = aiven_flink.testing.project
		  service_name   = aiven_flink.testing.service_name
		  integration_id = aiven_service_integration.flinkpg.integration_id
		  table_name     = var.sink_table_name
		  jdbc_table     = var.sink_jdbc_table_name
		  schema_sql     = <<EOF
		    cpu INT,
		    node INT,
		    occurred_at TIMESTAMP(3)
		  EOF
		}
		
		resource "aiven_flink_job" "testing" {
		  project      = aiven_flink_table.source.project
		  service_name = aiven_flink.testing.service_name
		  job_name     = var.job_name
		  table_ids = [
		    aiven_flink_table.source.table_id,
		    aiven_flink_table.sink.table_id
		  ]
		  statement = <<EOF
		    INSERT INTO ${aiven_flink_table.sink.table_name}
		    SELECT * FROM ${aiven_flink_table.source.table_name}
		    WHERE cpu > 75
		  EOF
		}`,

		projectName,
		flinkServiceName,
		kafkaServiceName,
		postgresServiceName,
		sourceTopicName,
		sourceTableName,
		sinkTableName,
		sinkJdbcTableName,
		jobName,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenFlinkJobsAndTableResourcesDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					// only check tables and jobs

					// source table
					resource.TestCheckResourceAttr("aiven_flink_table.source", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "service_name", flinkServiceName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "kafka_topic", sourceTopicName),
					resource.TestCheckResourceAttrSet("aiven_flink_table.source", "schema_sql"),

					// sink table
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "service_name", flinkServiceName),
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "jdbc_table", sinkJdbcTableName),
					resource.TestCheckResourceAttrSet("aiven_flink_table.sink", "schema_sql"),

					// job
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "service_name", flinkServiceName),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.0"),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.1"),
				),
			},
			{
				Config:       manifest,
				ImportState:  true,
				ResourceName: "aiven_flink_table.source",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["aiven_flink_table.source"]
					if !ok {
						return "", fmt.Errorf("expected resource 'aiven_flink_table.source' to be present in the state")
					}

					return rs.Primary.ID, nil
				},
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					if len(is) != 1 {
						return fmt.Errorf("expected only one instance to be imported, state: %#v", is)
					}

					tableId, ok := is[0].Attributes["table_id"]
					if !ok {
						return errors.New("expected the imported flink table to have table_id to be set")
					}

					expectedId := fmt.Sprintf("%s/%s/%s", projectName, flinkServiceName, tableId)

					if !strings.EqualFold(is[0].ID, expectedId) {
						return fmt.Errorf("expect the ID used in import statement to match '%s', but got: %s", expectedId, is[0].ID)
					}

					return nil
				},
			},
		},
	})
}

func TestAccAiven_flink_kafkaToKafka(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	flinkServiceName := fmt.Sprintf("test-acc-flink-%s", randString())
	kafkaServiceName := fmt.Sprintf("test-acc-flink-kafka-%s", randString())
	sourceTopicName := fmt.Sprintf("test-acc-flink-kafka-source-topic-%s", randString())
	sinkTopicName := fmt.Sprintf("test-acc-flink-kafka-sink-topic-%s", randString())
	sourceTableName := fmt.Sprintf("test_acc_flink_kafka_source_table_%s", randString())
	sinkTableName := fmt.Sprintf("test_acc_flink_kafka_sink_table_%s", randString())
	jobName := fmt.Sprintf("test_acc_flink_job_%s", randString())

	manifest := fmt.Sprintf(`
		variable "project_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_flink" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_kafka" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_topic_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_topic_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "job_name" {
		  type    = string
		  default = "%s"
		}
		
		resource "aiven_flink" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-4"
		  service_name = var.service_name_flink
		}
		
		resource "aiven_kafka" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-2"
		  service_name = var.service_name_kafka
		}
		
		resource "aiven_kafka_topic" "source" {
		  project      = aiven_kafka.testing.project
		  service_name = aiven_kafka.testing.service_name
		  topic_name   = var.source_topic_name
		  replication  = 2
		  partitions   = 2
		}
		
		resource "aiven_kafka_topic" "sink" {
		  project      = aiven_kafka.testing.project
		  service_name = aiven_kafka.testing.service_name
		  topic_name   = var.sink_topic_name
		  replication  = 2
		  partitions   = 2
		}
		
		resource "aiven_service_integration" "testing" {
		  project                  = aiven_flink.testing.project
		  integration_type         = "flink"
		  destination_service_name = aiven_flink.testing.service_name
		  source_service_name      = aiven_kafka.testing.service_name
		}
		
		resource "aiven_flink_table" "source" {
		  project              = aiven_flink.testing.project
		  service_name         = aiven_flink.testing.service_name
		  integration_id       = aiven_service_integration.testing.integration_id
		  table_name           = var.source_table_name
		  kafka_topic          = aiven_kafka_topic.source.topic_name
		  kafka_connector_type = "kafka"
		  kafka_value_format   = "json"
		  kafka_key_format     = "json"
		  kafka_key_fields     = ["cpu"]
		  kafka_startup_mode   = "earliest-offset"
		  schema_sql           = <<EOF
		    cpu INT,
		    node INT,
		    occurred_at TIMESTAMP(3) METADATA FROM 'timestamp',
		    WATERMARK FOR occurred_at AS occurred_at - INTERVAL '5' SECOND
		  EOF
		}
		
		resource "aiven_flink_table" "sink" {
		  project        = aiven_flink.testing.project
		  service_name   = aiven_flink.testing.service_name
		  integration_id = aiven_service_integration.testing.integration_id
		  table_name     = var.sink_table_name
		  kafka_topic    = aiven_kafka_topic.sink.topic_name
		  schema_sql     = <<EOF
		    cpu INT,
		    node INT,
		    occurred_at TIMESTAMP(3)
		  EOF
		}
		
		resource "aiven_flink_job" "testing" {
		  project      = aiven_flink.testing.project
		  service_name = aiven_flink.testing.service_name
		  job_name     = var.job_name
		  table_ids = [
		    aiven_flink_table.source.table_id,
		    aiven_flink_table.sink.table_id
		  ]
		  statement = <<EOF
		    INSERT INTO ${aiven_flink_table.sink.table_name}
		    SELECT * FROM ${aiven_flink_table.source.table_name}
		    WHERE cpu > 75
		  EOF
		}`,

		projectName,
		flinkServiceName,
		kafkaServiceName,
		sourceTopicName,
		sinkTopicName,
		sourceTableName,
		sinkTableName,
		jobName,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenFlinkJobsAndTableResourcesDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					// only check tables and jobs

					// source table
					resource.TestCheckResourceAttr("aiven_flink_table.source", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "service_name", flinkServiceName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "kafka_topic", sourceTopicName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "kafka_connector_type", "kafka"),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "kafka_key_format", "json"),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "kafka_value_format", "json"),
					resource.TestCheckResourceAttrSet("aiven_flink_table.source", "schema_sql"),

					// sink table
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "service_name", flinkServiceName),
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "kafka_topic", sinkTopicName),
					resource.TestCheckResourceAttrSet("aiven_flink_table.sink", "schema_sql"),

					// job
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "service_name", flinkServiceName),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.0"),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.1"),
				),
			},
		},
	})
}

func TestAccAiven_flink_kafkaToUpsertKafka(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	flinkServiceName := fmt.Sprintf("test-acc-flink-%s", randString())
	kafkaServiceName := fmt.Sprintf("test-acc-flink-kafka-%s", randString())
	sourceTopicName := fmt.Sprintf("test-acc-flink-kafka-source-topic-%s", randString())
	sinkTopicName := fmt.Sprintf("test-acc-flink-kafka-sink-topic-%s", randString())
	sourceTableName := fmt.Sprintf("test_acc_flink_kafka_source_table_%s", randString())
	sinkTableName := fmt.Sprintf("test_acc_flink_kafka_sink_table_%s", randString())
	jobName := fmt.Sprintf("test_acc_flink_job_%s", randString())

	manifest := fmt.Sprintf(`
		variable "project_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_flink" {
		  type    = string
		  default = "%s"
		}
		
		variable "service_name_kafka" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_topic_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_topic_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "source_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "sink_table_name" {
		  type    = string
		  default = "%s"
		}
		
		variable "job_name" {
		  type    = string
		  default = "%s"
		}
		
		resource "aiven_flink" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-4"
		  service_name = var.service_name_flink
		}
		
		resource "aiven_kafka" "testing" {
		  project      = var.project_name
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-2"
		  service_name = var.service_name_kafka
		}
		
		resource "aiven_kafka_topic" "source" {
		  project      = aiven_kafka.testing.project
		  service_name = aiven_kafka.testing.service_name
		  topic_name   = var.source_topic_name
		  replication  = 2
		  partitions   = 2
		}
		
		resource "aiven_kafka_topic" "sink" {
		  project      = aiven_kafka.testing.project
		  service_name = aiven_kafka.testing.service_name
		  topic_name   = var.sink_topic_name
		  replication  = 2
		  partitions   = 2
		}
		
		resource "aiven_service_integration" "testing" {
		  project                  = aiven_flink.testing.project
		  integration_type         = "flink"
		  destination_service_name = aiven_flink.testing.service_name
		  source_service_name      = aiven_kafka.testing.service_name
		}
		
		resource "aiven_flink_table" "source" {
		  project        = aiven_flink.testing.project
		  service_name   = aiven_flink.testing.service_name
		  integration_id = aiven_service_integration.testing.integration_id
		
		  upsert_kafka {
		    topic                = aiven_kafka_topic.source.topic_name
		    value_format         = "avro"
		    key_format           = "avro"
		    value_fields_include = "EXCEPT_KEY"
		  }
		  table_name = var.source_table_name
		
		  schema_sql = "cpu INT PRIMARY KEY"
		}
		
		resource "aiven_flink_table" "sink" {
		  project        = aiven_flink.testing.project
		  service_name   = aiven_flink.testing.service_name
		  integration_id = aiven_service_integration.testing.integration_id
		  table_name     = var.sink_table_name
		  upsert_kafka {
		    topic        = aiven_kafka_topic.source.topic_name
		    value_format = "avro"
		    key_format   = "avro"
		  }
		  schema_sql = "cpu INT PRIMARY KEY"
		}
		
		resource "aiven_flink_job" "testing" {
		  project      = aiven_flink.testing.project
		  service_name = aiven_flink.testing.service_name
		  job_name     = var.job_name
		  table_ids = [
		    aiven_flink_table.source.table_id,
		    aiven_flink_table.sink.table_id
		  ]
		  statement = <<EOF
		    INSERT INTO ${aiven_flink_table.sink.table_name}
		    SELECT * FROM ${aiven_flink_table.source.table_name}
		    WHERE cpu > 75
		  EOF
		}`,

		projectName,
		flinkServiceName,
		kafkaServiceName,
		sourceTopicName,
		sinkTopicName,
		sourceTableName,
		sinkTableName,
		jobName,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenFlinkJobsAndTableResourcesDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					// only check tables and jobs

					// source table
					resource.TestCheckResourceAttr("aiven_flink_table.source", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.source", "service_name", flinkServiceName),
					resource.TestCheckResourceAttrSet("aiven_flink_table.source", "schema_sql"),

					// sink table
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_table.sink", "service_name", flinkServiceName),
					resource.TestCheckResourceAttrSet("aiven_flink_table.sink", "schema_sql"),

					// job
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "project", projectName),
					resource.TestCheckResourceAttr("aiven_flink_job.testing", "service_name", flinkServiceName),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.0"),
					resource.TestCheckResourceAttrSet("aiven_flink_job.testing", "table_ids.1"),
				),
			},
		},
	})
}

func testAccCheckAivenFlinkJobsAndTableResourcesDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each job and table is destroyed
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aiven_flink_job":
			project, serviceName, jobId := schemautil.SplitResourceID3(rs.Primary.ID)

			_, err := c.Services.Get(project, serviceName)
			if err != nil {
				if aiven.IsNotFound(err) {
					continue
				}
				return err
			}

			r, err := c.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobId})
			if err != nil {
				if aiven.IsNotFound(err) {
					continue
				}
				return err
			}

			if r != nil {
				return fmt.Errorf("flink job (%s) still exists, id %s", jobId, rs.Primary.ID)
			}
		case "aiven_flink_table":
			project, serviceName, tableId := schemautil.SplitResourceID3(rs.Primary.ID)

			_, err := c.Services.Get(project, serviceName)
			if err != nil {
				if aiven.IsNotFound(err) {
					continue
				}
				return err
			}

			r, err := c.FlinkTables.Get(project, serviceName, aiven.GetFlinkTableRequest{TableId: tableId})
			if err != nil {
				if aiven.IsNotFound(err) {
					continue
				}
				return err
			}

			if r != nil {
				return fmt.Errorf("flink table (%s) still exists, id %s", tableId, rs.Primary.ID)
			}
		default:
			continue
		}
	}

	return nil
}
