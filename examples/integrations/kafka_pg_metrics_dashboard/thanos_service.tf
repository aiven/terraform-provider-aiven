resource "aiven_thanos" "example_thanos" {
  project                 = var.aiven_project_name
  cloud_name              = "google-europe-west1"
  plan                    = "startup-8"
  service_name            = "sample-thanos"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "11:00:00"
  thanos_user_config {
    ip_filter_object {
      network = "0.0.0.0/0"
    }
  }
}
