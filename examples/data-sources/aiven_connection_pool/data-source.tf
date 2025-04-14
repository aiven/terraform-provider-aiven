data "aiven_connection_pool" "main" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_pg.example_postgres.service_name
  pool_name    = "example-pool"
}
