resource "aiven_pg" "example_postgres" {
  project                 = var.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "example-postgres-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}