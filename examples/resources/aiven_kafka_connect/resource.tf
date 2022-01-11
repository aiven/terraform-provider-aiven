resource "aiven_kafka_connect" "kc1" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "my-kc1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = true
    }
  }
}
