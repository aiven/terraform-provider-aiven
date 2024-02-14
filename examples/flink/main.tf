terraform {
  required_version = ">=0.13"
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

provider "aiven" {
  api_token = var.aiven_api_token
}

data "aiven_project" "prj" {
  project = "flink-project"
}

resource "aiven_flink" "flink" {
  project      = data.aiven_project.prj.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "demo-flink"
}

resource "aiven_kafka" "kafka" {
  project      = data.aiven_project.prj.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "demo-kafka"
}

resource "aiven_service_integration" "flink_to_kafka" {
  project                  = data.aiven_project.prj.project
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

resource "aiven_flink_application" "app" {
  project      = data.aiven_project.prj.project
  service_name = aiven_flink.flink.service_name
  name         = "test-app"
}


resource "aiven_flink_application_version" "app_version" {
  project        = data.aiven_project.prj.project
  service_name   = aiven_flink.flink.service_name
  application_id = aiven_flink_application.app.application_id
  statement      = <<EOT
   INSERT INTO kafka_known_pizza SELECT * FROM kafka_pizza WHERE shop LIKE '%Luigis Pizza%'
  EOT
  sink {
    create_table   = <<EOT
    CREATE TABLE kafka_known_pizza (
        shop STRING,
        name STRING
    ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'sink_topic',
        'value.format' = 'json'
    )
  EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
  source {
    create_table   = <<EOT
    CREATE TABLE kafka_pizza (
        shop STRING,
        name STRING
    ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'source_topic',
        'value.format' = 'json'
    )
    EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
}

resource "aiven_flink_application_deployment" "deployment" {
  project        = data.aiven_project.prj.project
  service_name   = aiven_flink.flink.service_name
  application_id = aiven_flink_application.app.application_id
  version_id     = aiven_flink_application_version.app_version.application_version_id
}
