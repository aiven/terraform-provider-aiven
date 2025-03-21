resource "aiven_service_integration_endpoint" "autoscaler_endpoint" {
  project       = var.aiven_project.example_project.project
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
  project                 = var.aiven_project.example_project.project
  integration_type        = "autoscaler"
  source_service_name     = aiven_pg.example_postgres.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.autoscaler_endpoint.id
}