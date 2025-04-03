# Kafka service
resource "aiven_kafka" "kafka_service" {
  project                 = var.aiven_project_name
  service_name            = var.kafka_service_name
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
