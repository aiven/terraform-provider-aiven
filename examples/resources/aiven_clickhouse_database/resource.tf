resource "aiven_clickhouse_database" "clickhouse_db" {
  project      = aiven_clickhouse.ch.project
  service_name = aiven_clickhouse.ch.service_name
  name         = "my-ch-db"
}
