resource "aiven_elasticsearch" "es1" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "my-es1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"

    elasticsearch_user_config {
        elasticsearch_version = 7

        kibana {
            enabled = true
            elasticsearch_request_timeout = 30000
        }

        public_access {
            elasticsearch = true
            kibana = true
        }
    }
}
