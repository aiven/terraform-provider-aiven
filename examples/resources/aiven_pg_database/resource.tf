resource "aiven_pg_database" "main" {
  project       = aiven_project.example_project.project
  service_name  = aiven_pg.example_postgres.service_name
  database_name = "example-database"
}
