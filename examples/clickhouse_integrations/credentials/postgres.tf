# Postgres service based in GCP US East
resource "aiven_pg" "external_postgres" {
  project      = var.aiven_project_name
  service_name = var.external_postgres_service_name
  cloud_name   = "google-us-east4"
  plan         = "business-8" # Primary and read replica
}

resource "aiven_service_integration_endpoint" "external_postgres" {
  project       = var.aiven_project_name
  endpoint_name = "external-postgresql"
  endpoint_type = "external_postgresql"

  external_postgresql {
    host     = aiven_pg.external_postgres.service_host
    port     = aiven_pg.external_postgres.service_port
    username = aiven_pg.external_postgres.service_username
    password = aiven_pg.external_postgres.service_password
  }
}

