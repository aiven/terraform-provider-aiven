resource "aiven_grafana" "samplegrafana" {
  project      = var.aiven_project_name
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "samplegrafana"
  grafana_user_config {
    ip_filter_object {
      network = "0.0.0.0/0"
    }
  }
}
