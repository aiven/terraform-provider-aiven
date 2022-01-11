resource "aiven_influxdb" "inf1" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "my-inf1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  influxdb_user_config {
    public_access {
      influxdb = true
    }
  }
}
