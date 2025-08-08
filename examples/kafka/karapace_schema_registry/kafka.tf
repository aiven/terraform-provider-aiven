resource "aiven_kafka" "kafka_karapace_schema_registry" {
  project                 = var.aiven_project
  cloud_name              = "google-northamerica-northeast1"
  plan                    = "business-4"
  service_name            = var.kafka_service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    // Enables Karapace schema registry and REST
    schema_registry = true
    kafka_rest      = true
    // Enables automatic topic creation
    kafka {
      auto_create_topics_enable = true
    }
  }
}

resource "aiven_kafka_topic" "source" {
  project      = var.aiven_project
  service_name = aiven_kafka.demo-kafka.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}
