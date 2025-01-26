resource "aiven_kafka_mirrormaker" "example_mirrormaker" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "example-mirrormaker-service"

  kafka_mirrormaker_user_config {
    ip_filter = ["0.0.0.0/0"]

    kafka_mirrormaker {
      refresh_groups_interval_seconds = 600
      refresh_topics_enabled          = true
      refresh_topics_interval_seconds = 600
    }
  }
}
