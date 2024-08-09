resource "aiven_flink" "example_flink" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "example-flink-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  flink_user_config {
    flink_version = 1.19
  }
}
