# ClickHouse service based in the same region
resource "aiven_clickhouse" "clickhouse" {
  project                 = aiven_project.clickhouse_postgres_source.project
  service_name            = "clickhouse-gcp-us"
  cloud_name              = "google-us-east4"
  plan                    = "startup-beta-16"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

}

# ClickHouse service integration for the PostgreSQL service as source
# exposing three databases in the public schema
resource "aiven_service_integration" "clickhouse_postgres_source" {
  project                  = aiven_project.clickhouse_postgres_source.project
  integration_type         = "clickhouse_postgresql"
  source_service_name      = aiven_pg.postgres.service_name
  destination_service_name = aiven_clickhouse.clickhouse.service_name
  clickhouse_postgresql_user_config {
    databases {
      database = aiven_pg_database.suppliers_dims.database_name
      schema   = "public"
    }
    databases {
      database = aiven_pg_database.inventory_facts.database_name
      schema   = "public"
    }
    databases {
      database = aiven_pg_database.order_events.database_name
      schema   = "public"
    }
  }
}
