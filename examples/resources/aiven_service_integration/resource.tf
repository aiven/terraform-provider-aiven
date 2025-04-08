# Integrate Kafka and Thanos services for metrics
resource "aiven_service_integration" "example_integration" {
  project                  = data.aiven_project.example_project.project
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.example_kafka.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}

# Use disk autoscaler with a PostgreSQL service
resource "aiven_service_integration_endpoint" "autoscaler_endpoint" {
  project                  = data.aiven_project.example_project.project
  endpoint_name = "disk-autoscaler-200GiB"
  endpoint_type = "autoscaler"

  autoscaler_user_config {
    autoscaling {
      cap_gb = 200
      type   = "autoscale_disk"
    }
  }
}

resource "aiven_service_integration" "autoscaler_integration" {
  project                  = data.aiven_project.example_project.project
  integration_type         = "autoscaler"
  source_service_name     = aiven_pg.example_pg.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.autoscaler_endpoint.id
}