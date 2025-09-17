# Project
resource "aiven_project" "clickhouse_kafka_source" {
  project   = var.aiven_project
  parent_id = var.aiven_organization
}
