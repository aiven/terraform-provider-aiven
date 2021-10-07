resource "aiven_flink" "flink" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "business-4"
    service_name = "my-flink"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    flink_user_config {
        flink_version = 1.13
    }
}
