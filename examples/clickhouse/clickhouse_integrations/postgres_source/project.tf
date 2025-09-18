// Sample project name
resource "aiven_project" "clickhouse_postgres_source" {
  project   = var.aiven_project
  parent_id = var.aiven_organization
}
