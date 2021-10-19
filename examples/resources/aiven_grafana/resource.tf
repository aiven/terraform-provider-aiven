resource "aiven_grafana" "gr1" {
    project = data.aiven_project.ps1.project
    cloud_name = "google-europe-west1"
    plan = "startup-1"
    service_name = "my-gr1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    grafana_user_config {
        alerting_enabled = true
        
        public_access {
            grafana = true
        }
    }
}
