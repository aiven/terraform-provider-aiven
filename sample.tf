variable "aiven_api_token" {}
variable "aiven_card_id" {}
variable "aiven_project_name" {}

terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = "2.X.X"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
  api_token = "${var.aiven_api_token}"
}

# Project
resource "aiven_project" "sample" {
  project = "${var.aiven_project_name}"
  card_id = "${var.aiven_card_id}"
}

# Kafka service
resource "aiven_service" "samplekafka" {
  project                 = "${aiven_project.sample.project}"
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "samplekafka"
  service_type            = "kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka_connect = true
    kafka_rest    = true
    kafka_version = "2.6"
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

# Topic for Kafka
resource "aiven_kafka_topic" "sample_topic" {
  project         = "${aiven_project.sample.project}"
  service_name    = "${aiven_service.samplekafka.service_name}"
  topic_name      = "sample_topic"
  partitions      = 3
  replication     = 2
  retention_bytes = 1000000000
}

# User for Kafka
resource "aiven_service_user" "kafka_a" {
  project      = "${aiven_project.sample.project}"
  service_name = "${aiven_service.samplekafka.service_name}"
  username     = "kafka_a"
}

# ACL for Kafka
resource "aiven_kafka_acl" "sample_acl" {
  project      = "${aiven_project.sample.project}"
  service_name = "${aiven_service.samplekafka.service_name}"
  username     = "kafka_*"
  permission   = "read"
  topic        = "*"
}

# InfluxDB service
resource "aiven_service" "sampleinflux" {
  project                 = "${aiven_project.sample.project}"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "sampleinflux"
  service_type            = "influxdb"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "11:00:00"
  influxdb_user_config {
    ip_filter = ["0.0.0.0/0"]
  }
}

# Send metrics from Kafka to InfluxDB
resource "aiven_service_integration" "samplekafka_metrics" {
  project                  = "${aiven_project.sample.project}"
  integration_type         = "metrics"
  source_service_name      = "${aiven_service.samplekafka.service_name}"
  destination_service_name = "${aiven_service.sampleinflux.service_name}"
}

# PostreSQL service
resource "aiven_service" "samplepg" {
  project                 = "${aiven_project.sample.project}"
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "samplepg"
  service_type            = "pg"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "12:00:00"
  pg_user_config {
    pg {
      idle_in_transaction_session_timeout = 900
    }
    pg_version = "10"
  }
}

# Send metrics from PostgreSQL to InfluxDB
resource "aiven_service_integration" "samplepg_metrics" {
  project                  = "${aiven_project.sample.project}"
  integration_type         = "metrics"
  source_service_name      = "${aiven_service.samplepg.service_name}"
  destination_service_name = "${aiven_service.sampleinflux.service_name}"
}

# PostgreSQL database
resource "aiven_database" "sample_db" {
  project       = "${aiven_project.sample.project}"
  service_name  = "${aiven_service.samplepg.service_name}"
  database_name = "sample_db"
}

# PostgreSQL user
resource "aiven_service_user" "sample_user" {
  project      = "${aiven_project.sample.project}"
  service_name = "${aiven_service.samplepg.service_name}"
  username     = "sampleuser"
}

# PostgreSQL connection pool
resource "aiven_connection_pool" "sample_pool" {
  project       = "${aiven_project.sample.project}"
  service_name  = "${aiven_service.samplepg.service_name}"
  database_name = "${aiven_database.sample_db.database_name}"
  pool_name     = "samplepool"
  username      = "${aiven_service_user.sample_user.username}"
}

# Grafana service
resource "aiven_service" "samplegrafana" {
  project      = "${aiven_project.sample.project}"
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "samplegrafana"
  service_type = "grafana"
  grafana_user_config {
    ip_filter = ["0.0.0.0/0"]
  }
}

# Dashboards for Kafka and PostgreSQL services
resource "aiven_service_integration" "samplegrafana_dashboards" {
  project                  = "${aiven_project.sample.project}"
  integration_type         = "dashboard"
  source_service_name      = "${aiven_service.samplegrafana.service_name}"
  destination_service_name = "${aiven_service.sampleinflux.service_name}"
}
