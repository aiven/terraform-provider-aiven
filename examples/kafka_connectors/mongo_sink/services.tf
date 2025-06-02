# Kafka service
resource "aiven_kafka" "kafka-service1" {
  project                 = aiven_project.kafka-con-project1.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "kafka-service1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_version = "3.8"

    // Enables Kafka Connectors
    kafka_connect = true

    // Enable Kafka Schema Registry
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

# Kafka topic
resource "aiven_kafka_topic" "kafka-topic1" {
  project      = aiven_project.kafka-con-project1.project
  service_name = aiven_kafka.kafka-service1.service_name
  topic_name   = "test-kafka-topic1"
  partitions   = 3
  replication  = 2
}
