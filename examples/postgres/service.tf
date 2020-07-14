# Grafana service
resource "aiven_service" "avn-pg" {
  project      = var.avn_project_id
  cloud_name   = "google-europe-west1"
  plan         = "premium-8"
  service_name = "postgres-eu"
  service_type = "pg"
}
