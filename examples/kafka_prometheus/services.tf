resource "aiven_kafka" "kafka" {
  project                 = var.avn_project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.kafka_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_rest      = true
    kafka_connect   = true
    schema_registry = true
    kafka_version = "3.2"

    public_access {
      kafka_rest    = true
      kafka_connect = true
      prometheus    = true
    }
  }
}

resource "aiven_service_integration_endpoint" "endpoint" {
  project       = aiven_kafka.kafka.project
  endpoint_name = var.prometheus_endpoint_name
  endpoint_type = "prometheus"

  prometheus_user_config {
    basic_auth_username = var.prometheus_username
    basic_auth_password = var.prometheus_password
  }
}

resource "aiven_service_integration" "integration" {
  project                 = aiven_kafka.kafka.project
  source_service_name     = aiven_kafka.kafka.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.endpoint.id
  integration_type        = "prometheus"
}
