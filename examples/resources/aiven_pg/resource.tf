resource "aiven_pg" "pg" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "my-pg1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"

    pg_user_config {
        pg_version = 11

        public_access {
            pg = true
            prometheus = false
        }

        pg {
            idle_in_transaction_session_timeout = 900
            log_min_duration_statement = -1
        }
    }

    timeouts {
        create = "20m"
        update = "15m"
    }
}
