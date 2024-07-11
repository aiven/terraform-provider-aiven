# Create a Kafka service.
resource "aiven_kafka" "example_kafka" {
  project      = data.aiven_project.example_project.project
  service_name = "example-kafka-service"
  cloud_name   = "google-europe-west1"
  plan         = "startup-2"
}

# Create a Kafka Connect service.
resource "aiven_kafka_connect" "example_kafka_connect" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "example-connect-service"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = true
    }
  }
}

# Integrate the Kafka and Kafka Connect services.
resource "aiven_service_integration" "kafka_connect_integration" {
  project                  = data.aiven_project.example_project.project
  integration_type         = "kafka_connect"
  source_service_name      = aiven_kafka.example_kafka.service_name
  destination_service_name = aiven_kafka_connect.example_kafka_connect.service_name

  kafka_connect_user_config {
    kafka_connect {
      group_id             = "connect"
      status_storage_topic = "__connect_status"
      offset_storage_topic = "__connect_offsets"
    }
  }
}
