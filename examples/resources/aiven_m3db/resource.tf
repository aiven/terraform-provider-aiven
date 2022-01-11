resource "aiven_m3db" "m3" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-8"
  service_name            = "my-m3db"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  m3db_user_config {
    m3db_version = 0.15

    namespaces {
      name = "my-ns1"
      type = "unaggregated"
    }
  }
}
