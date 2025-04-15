# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# Aiven for Metrics (Thanos) service
resource "aiven_thanos" "example_thanos" {
  project      = data.aiven_project.main.project
  plan         = "startup-4"
  cloud_name   = "google-europe-west1"
  service_name = var.thanos_service_name

  thanos_user_config {
    compactor {
      retention_days = "30"
    }
    object_storage_usage_alert_threshold_gb = "10"
  }
}
