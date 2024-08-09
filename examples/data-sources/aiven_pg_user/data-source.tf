data "aiven_pg_user" "example_user" {
  service_name = aiven_pg.example_postgres.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-service-user"
}