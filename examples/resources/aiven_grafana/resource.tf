resource "aiven_grafana" "example_grafana" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "example-grafana-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  grafana_user_config {
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}
