data "aiven_service_component" "grafana_public" {
  project      = aiven_grafana.grafana_service.project
  service_name = aiven_grafana.grafana_service.service_name
  component    = "grafana"
  route        = "public"

  depends_on = [
    aiven_grafana.grafana_service
  ]
}

output "grafana_public" {
  value = "${data.aiven_service_component.grafana_public.host}:${data.aiven_service_component.grafana_public.port}"
}
