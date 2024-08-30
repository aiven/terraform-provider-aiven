resource "aiven_m3db" "example_m3db" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-8"
  service_name            = "example-m3db-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  m3db_user_config {
    m3db_version = 1.1

    namespaces {
      name = "example-namespace"
      type = "unaggregated"
    }
  }
}
