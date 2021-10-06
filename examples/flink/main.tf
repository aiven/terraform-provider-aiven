terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}

variable "aiven_api_token" {
  type = string
}

variable "aiven_card_id" {
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
  project = aiven_project.project.project
  cloud_name = "google-europe-west1"
  plan = "business-8"
  service_name = "demo-flink"
}

resource "aiven_kafka" "kafka" {
  project = aiven_project.project.project
  cloud_name = "google-europe-west1"
  plan = "business-8"
  service_name = "demo-kafka"
}

resource "aiven_service_integration" "flink_to_kafka" {
  project = aiven_project.project.project
  integration_type = "flink"
  destination_service_name = aiven_flink.flink.service_name
  source_service_name = aiven_kafka.kafka.service_name
}

resource "aiven_kafka_topic" "source" {
  project = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions = 2
  replication = 3
  topic_name = "source_topic"
}

resource "aiven_kafka_topic" "sink" {
  project = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions = 2
  replication = 3
  topic_name = "sink_topic"
}

resource "aiven_flink_table" "source" {
  project = aiven_flink.flink.project
  service_name = aiven_flink.flink.service_name
  integration_id = aiven_service_integration.flink_to_kafka.integration_id
  table_name = "source_table"
  kafka_topic = aiven_kafka_topic.source.topic_name
  partitioned_by = "node"
  schema_sql = <<EOF
    `cpu` INT, 
    `node` INT, 
    `occurred_at` TIMESTAMP(3) METADATA FROM 'timestamp', 
    `cpu_in_mb` AS `cpu` * 4 * 1024, 
    WATERMARK FOR `occurred_at` AS `occurred_at` - INTERVAL '5' SECOND
  EOF
}

resource "aiven_flink_table" "sink" {
  project = aiven_flink.flink.project
  service_name = aiven_flink.flink.service_name
  integration_id = aiven_service_integration.flink_to_kafka.integration_id
  table_name = "sink_table"
  kafka_topic = aiven_kafka_topic.sink.topic_name
  schema_sql = <<EOF
    `cpu` INT, 
    `node` INT, 
    `occurred_at` TIMESTAMP(3), 
    `cpu_in_mb` FLOAT
  EOF
}

resource "aiven_flink_job" "flink_job" {
  project = aiven_flink.flink.project
  service_name = aiven_flink.flink.service_name
  job_name = "my_job"
  tables = [
    aiven_flink_table.source.table_name,
    aiven_flink_table.sink.table_name
  ]
  statement =<<EOF
    INSERT INTO ${aiven_flink_table.sink.table_name} 
    SELECT * FROM ${aiven_flink_table.source.table_name} 
    WHERE `cpu` > 70
  EOF
}

