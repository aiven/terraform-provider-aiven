data "aiven_clickhouse_database" "example" {
  project      = "my-project"
  service_name = "my-clickhouse"
  name         = "testdb"
}
