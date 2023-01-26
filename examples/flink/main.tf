terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

variable "aiven_api_token" {
  type = string
}

variable "aiven_card_id" {
  type = string
}

provider "aiven" {
  api_token = var.aiven_api_token
}

resource "aiven_project" "project" {
  project = "flink-project"
  card_id = var.aiven_card_id
}

resource "aiven_flink" "flink" {
  project      = aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "demo-flink"
}

resource "aiven_kafka" "kafka" {
  project      = aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "demo-kafka"
}

resource "aiven_service_integration" "flink_to_kafka" {
  project                  = aiven_project.project.project
  integration_type         = "flink"
  destination_service_name = aiven_flink.flink.service_name
  source_service_name      = aiven_kafka.kafka.service_name
}

resource "aiven_kafka_topic" "source" {
  project      = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions   = 2
  replication  = 3
  topic_name   = "source_topic"
}

resource "aiven_kafka_topic" "sink" {
  project      = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions   = 2
  replication  = 3
  topic_name   = "sink_topic"
}

