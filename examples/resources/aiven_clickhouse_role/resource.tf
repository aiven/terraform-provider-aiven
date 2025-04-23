resource "aiven_clickhouse_role" "example_role" {
  service_name = aiven_clickhouse.example_clickhouse.service_name
  project      = data.aiven_project.example_project.project
  role         = "writer"
}
