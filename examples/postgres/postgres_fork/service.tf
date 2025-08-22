# Source PostgreSQL service
resource "aiven_pg" "original_postgres_service" {
  project      = var.aiven_project
  cloud_name   = "google-europe-west1"
  plan         = "startup-8"
  service_name = var.source_pg_name
}

# Forked PostgreSQL service
resource "aiven_pg" "postgres_fork" {
  project      = var.aiven_project
  cloud_name   = "eu-central-1"
  plan         = "startup-8"
  service_name = var.pg_fork_name
  pg_user_config {
    service_to_fork_from = aiven_pg.original_postgres_service.service_name
  }
}
