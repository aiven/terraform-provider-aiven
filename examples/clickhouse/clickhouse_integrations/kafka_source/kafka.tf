# Kafka service on GCP eu-west
resource "aiven_kafka" "kafka" {
  project                 = aiven_project.clickhouse_kafka_source.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "kafka-gcp-eu"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  # Enable kafka REST to view and send messages from the Console
  kafka_user_config {
    kafka_rest = true

    public_access {
      kafka_rest = true
    }
  }
}

# Kafka topic used to ingest edge measurements from the IoT devices fleet
resource "aiven_kafka_topic" "edge_measurements" {
  project                = aiven_project.clickhouse_kafka_source.project
  service_name           = aiven_kafka.kafka.service_name
  topic_name             = "edge-measurements"
  partitions             = 50
  replication            = 3
  termination_protection = false

  config {
    flush_ms        = 10
    cleanup_policy  = "delete"
    retention_bytes = "134217728" # 10 GiB
    retention_ms    = "604800000" # 1 week
  }
}
