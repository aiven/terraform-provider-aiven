package flink_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
  }
}

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
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
  }
}

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
		serviceName,
	)
	resourceName := "aiven_flink.bar"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
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

func TestAccAiven_flink_kafkaToPG(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	flinkServiceName := fmt.Sprintf("test-acc-flink-%s", randString())
	kafkaServiceName := fmt.Sprintf("test-acc-flink-kafka-%s", randString())
	postgresServiceName := fmt.Sprintf("test-acc-flink-postgres-%s", randString())
	sourceTopicName := fmt.Sprintf("test-acc-flink-kafka-source-topic-%s", randString())
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
`, projectName,
		flinkServiceName,
		kafkaServiceName,
		postgresServiceName,
		sourceTopicName,
	)

	resourceName := "aiven_flink.testing"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", flinkServiceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
				),
			},
		},
	})
}

func TestAccAiven_flink_kafkaToOS(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }
	flinkServiceName := fmt.Sprintf("test-acc-flink-%s", randString())
	kafkaServiceName := fmt.Sprintf("test-acc-flink-kafka-%s", randString())
	openSearchServiceName := fmt.Sprintf("test-acc-flink-opensearch-%s", randString())
	sourceTopicName := fmt.Sprintf("test-acc-flink-kafka-source-topic-%s", randString())
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

variable "service_name_os" {
  type    = string
  default = "%s"
}

variable "source_topic_name" {
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

resource "aiven_opensearch" "testing" {
  project      = var.project_name
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = var.service_name_os
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

resource "aiven_service_integration" "flinkos" {
  project                  = aiven_flink.testing.project
  integration_type         = "flink"
  destination_service_name = aiven_flink.testing.service_name
  source_service_name      = aiven_opensearch.testing.service_name
}
`, projectName,
		flinkServiceName,
		kafkaServiceName,
		openSearchServiceName,
		sourceTopicName,
	)

	resourceName := "aiven_flink.testing"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", flinkServiceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
				),
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


`, projectName,
		flinkServiceName,
		kafkaServiceName,
		sourceTopicName,
		sinkTopicName,
	)

	resourceName := "aiven_flink.testing"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", flinkServiceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
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
`, projectName,
		flinkServiceName,
		kafkaServiceName,
		sourceTopicName,
		sinkTopicName,
	)

	resourceName := "aiven_flink.testing"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", flinkServiceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
				),
			},
		},
	})
}
