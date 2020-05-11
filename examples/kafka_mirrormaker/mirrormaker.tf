resource "aiven_service" "mm" {
  project = aiven_project.kafka-mm-project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "mm"
  service_type = "kafka_mirrormaker"

  kafka_mirrormaker_user_config {
    ip_filter = ["0.0.0.0/0"]

    kafka_mirrormaker {
      refresh_groups_interval_seconds = 600
      refresh_topics_enabled = true
      refresh_topics_interval_seconds = 600
    }
  }
}

resource "aiven_service_integration" "i1" {
  project = aiven_project.kafka-mm-project1.project
  integration_type = "kafka_mirrormaker"
  source_service_name = aiven_service.source.service_name
  destination_service_name = aiven_service.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = "source"
  }
}

resource "aiven_service_integration" "i2" {
  project = aiven_project.kafka-mm-project1.project
  integration_type = "kafka_mirrormaker"
  source_service_name = aiven_service.target.service_name
  destination_service_name = aiven_service.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = "target"
  }
}