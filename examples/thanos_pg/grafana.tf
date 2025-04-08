# Grafana dashboard for Thanos
resource "aiven_grafana" "thanos_dashboard" {
  project      = data.aiven_project.main.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = var.grafana_service_name

  grafana_user_config {
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}

resource "aiven_service_integration" "thanos_grafana_integration" {
  project                  = data.aiven_project.main.project
  integration_type         = "dashboard"
  source_service_name      = aiven_grafana.thanos_dashboard.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}
