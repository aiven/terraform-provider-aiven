# Datadog endpoint
resource "aiven_service_integration_endpoint" "example_endpoint" {
  project       = data.aiven_project.example_project.project
  endpoint_name = "Datadog endpoint"
  endpoint_type = "datadog"

  datadog_user_config {
    datadog_api_key = var.datadog_api_key
  }
}

# Disk autoscaler endpoint
resource "aiven_service_integration_endpoint" "autoscaler_endpoint" {
  project       = data.aiven_project.example_project.project
  endpoint_name = "disk-autoscaler-200GiB"
  endpoint_type = "autoscaler"

  autoscaler_user_config {
    autoscaling {
      cap_gb = 200
      type   = "autoscale_disk"
    }
  }
}
