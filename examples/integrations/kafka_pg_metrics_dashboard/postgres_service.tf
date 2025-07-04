# PostgreSQL service
resource "aiven_pg" "samplepg" {
  project                 = var.aiven_project_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "sample-pg"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "12:00:00"
  pg_user_config {
    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

# PostgreSQL database
resource "aiven_pg_database" "sample_db" {
  project       = var.aiven_project_name
  service_name  = aiven_pg.samplepg.service_name
  database_name = "sample_db"
}

# PostgreSQL service user
resource "aiven_pg_user" "sample_user" {
  project      = var.aiven_project_name
  service_name = aiven_pg.samplepg.service_name
  username     = "sampleuser"
}

# PostgreSQL connection pool
resource "aiven_connection_pool" "sample_pool" {
  project       = var.aiven_project_name
  service_name  = aiven_pg.samplepg.service_name
  database_name = aiven_pg_database.sample_db.database_name
  pool_name     = "samplepool"
  username      = aiven_pg_user.sample_user.username

  depends_on = [
    aiven_pg_database.sample_db,
  ]
}
