# Grafana service
resource "aiven_grafana" "grafana_service" {
  project      = var.aiven_project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "example-grafana-service"
  grafana_user_config {
    public_access {
      grafana = true
    }
  }
}
