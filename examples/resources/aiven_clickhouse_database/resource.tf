resource "aiven_clickhouse_database" "example" {
  project      = "my-project" // Force new
  service_name = "my-clickhouse" // Force new
  name         = "testdb" // Force new
}
