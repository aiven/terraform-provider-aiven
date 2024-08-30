resource "aiven_connection_pool" "main" {
  project       = data.aiven_project.example_project.project
  service_name  = aiven_pg.example_postgres.service_name
  database_name = aiven_pg_database.main.database_name
  pool_mode     = "transaction"
  pool_name     = "example-pool"
  pool_size     = 10
  username      = aiven_pg_user.example_user.username
}
