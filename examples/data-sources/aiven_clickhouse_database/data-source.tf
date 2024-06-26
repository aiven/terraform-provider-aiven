data "aiven_clickhouse_database" "example_clickhouse_db" {
  project      = data.aiven_clickhouse.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  name         = "example-database"
}