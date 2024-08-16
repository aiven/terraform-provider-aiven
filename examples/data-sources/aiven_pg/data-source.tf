data "aiven_pg" "example_postgres" {
  project      = data.aiven_project.example_project.project
  service_name = "example-postgres-service"
}

