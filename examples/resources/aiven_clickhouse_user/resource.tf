resource "aiven_clickhouse_user" "example_user" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_service.service_name
  username     = "analyst"
}
