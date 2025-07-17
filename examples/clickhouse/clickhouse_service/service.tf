# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# ClickHouse service
resource "aiven_clickhouse" "dev" {
  project                 = data.aiven_project.main.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = var.clickhouse_service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

# ClickHouse database that can be used to write the raw data
resource "aiven_clickhouse_database" "iot_analytics" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  name         = "iot_analytics"
}
