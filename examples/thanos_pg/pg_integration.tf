# PostgreSQL service
resource "aiven_pg" "example_postgres" {
  project      = data.aiven_project.main.project
  cloud_name   = "google-europe-west1"
  service_name = var.postgres_service_name
  plan         = "startup-4"
}

# Integration to send Postgres data to Thanos
resource "aiven_service_integration" "thanos_pg_integration" {
  project                  = data.aiven_project.main.project
  integration_type         = "metrics"
  source_service_name      = aiven_pg.example_postgres.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}
