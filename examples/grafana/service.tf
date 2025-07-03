# Grafana service
resource "aiven_grafana" "grafana-service1" {
  project      = aiven_project.project1.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "samplegrafana"
  grafana_user_config {
    public_access {
      grafana = true
    }
  }
}

data "aiven_service_component" "grafana_public" {
  project      = aiven_grafana.grafana-service1.project
  service_name = aiven_grafana.grafana-service1.service_name
  component    = "grafana"
  route        = "public"

  depends_on = [
    aiven_grafana.grafana-service1
  ]
}

output "grafana_public" {
  value = "${data.aiven_service_component.grafana_public.host}:${data.aiven_service_component.grafana_public.port}"
}
