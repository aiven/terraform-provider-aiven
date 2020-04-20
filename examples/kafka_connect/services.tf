# Kafka connect service
resource "aiven_service" "kafka_connect1" {
  project = aiven_project.kafka-con-project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "kafka-connect1"
  service_type = "kafka_connect"
  maintenance_window_dow = "monday"
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
