resource "aiven_kafka" "kafka1" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "my-kafka1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_rest      = true
    kafka_connect   = true
    schema_registry = true
    kafka_version   = "3.5"

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }

    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}
