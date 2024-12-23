data "aiven_clickhouse" "example_clickhouse" {
  project      = data.aiven_project.example_project.project
  service_name = "example-clickhouse-service"
}
