resource "aiven_thanos" "example_thanos" {
  project      = data.aiven_project.example_project.project
  plan         = "startup-4"
  cloud_name   = "google-europe-west1"
  service_name = "example-thanos-service"

  thanos_user_config {
    compactor {
      retention_days = "30"
    }
    object_storage_usage_alert_threshold_gb = "10"
  }
}
