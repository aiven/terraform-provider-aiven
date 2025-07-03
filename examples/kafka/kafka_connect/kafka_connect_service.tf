# Kafka connect service
resource "aiven_kafka_connect" "kafka_connect" {
  project                 = var.aiven_project_name
  service_name            = var.kafka_connect_name
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = true
    }
  }
}

# Kafka Connect service integration with Kafka service
resource "aiven_service_integration" "kafka_integration" {
  project                  = var.aiven_project_name
  integration_type         = "kafka_connect"
  source_service_name      = aiven_kafka.kafka_service.service_name
  destination_service_name = aiven_kafka_connect.kafka_connect.service_name

  kafka_connect_user_config {
    kafka_connect {
      group_id             = "connect"
      status_storage_topic = "__connect_status"
      offset_storage_topic = "__connect_offsets"
    }
  }
}
