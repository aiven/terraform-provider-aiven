resource "aiven_clickhouse" "example_clickhouse" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "example-clickhouse-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "example_db" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  name         = "example-database"
}
