resource "aiven_m3aggregator" "m3a" {
    project = data.aiven_project.foo.project
    cloud_name = "google-europe-west1"
    plan = "business-8"
    service_name = "my-m3a"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    m3aggregator_user_config {
      m3aggregator_version = 0.15
    }
}
