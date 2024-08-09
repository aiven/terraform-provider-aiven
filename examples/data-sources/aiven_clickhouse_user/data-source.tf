data "aiven_clickhouse_user" "example_user" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  username     = "analyst"
}
