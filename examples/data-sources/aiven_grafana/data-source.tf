data "aiven_grafana" "example_grafana" {
  project      = data.aiven_project.example_project.project
  service_name = "example-grafana-service"
}
