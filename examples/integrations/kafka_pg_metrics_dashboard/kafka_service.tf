# Kafka service
resource "aiven_kafka" "samplekafka" {
  project                 = var.aiven_project_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "sample-kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka_connect = true
    kafka_rest    = true
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

# Kafka topic
resource "aiven_kafka_topic" "sample_topic" {
  project      = var.aiven_project_name
  service_name = aiven_kafka.samplekafka.service_name
  topic_name   = "sample_topic"
  partitions   = 3
  replication  = 2
  config {
    retention_bytes = 1000000000
  }
}

# Kafka service user
resource "aiven_kafka_user" "kafka_a" {
  project      = var.aiven_project_name
  service_name = aiven_kafka.samplekafka.service_name
  username     = "kafka_a"
}

# Kafka ACL for the service user
resource "aiven_kafka_acl" "sample_acl" {
  project      = var.aiven_project_name
  service_name = aiven_kafka.samplekafka.service_name
  username     = "kafka_*"
  permission   = "read"
  topic        = "*"
}
