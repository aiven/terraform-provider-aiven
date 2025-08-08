# Kafka service with tiered storage enabled
resource "aiven_kafka" "kafka" {
  project                 = var.aiven_project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.kafka_service_name

  kafka_user_config {
    // Enables tiered storage
    tiered_storage {
        enabled = true
    }
  }
}

# Kafka Topic
resource "aiven_kafka_topic" "sample_topic" {
  project      = var.aiven_project
  service_name = aiven_kafka.kafka.service_name
  topic_name   = "sample-topic-with-tiered-storage"
  partitions   = 3
  replication  = 2
  config {
    // Enables tiered storage for the topic
    remote_storage_enable = true
    // Configures the retention time for the topic
    local_retention_ms = 300000 # 5 minutes
    segment_bytes = 1000000 # 1 Mb
  }
}