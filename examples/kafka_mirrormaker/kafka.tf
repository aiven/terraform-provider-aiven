resource "aiven_service" "source" {
  project = aiven_project.kafka-mm-project1.project
  cloud_name = "google-europe-west1"
  plan = "business-4"
  service_name = "source"
  service_type = "kafka"
  maintenance_window_dow = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_version = "2.4"
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "source" {
  project = aiven_project.kafka-mm-project1.project
  service_name = aiven_service.source.service_name
  topic_name = "topic-a"
  partitions = 3
  replication = 2
}

resource "aiven_service" "target" {
  project = aiven_project.kafka-mm-project1.project
  cloud_name = "google-europe-west1"
  plan = "business-4"
  service_name = "target"
  service_type = "kafka"
  maintenance_window_dow = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_version = "2.4"
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "target" {
  project = aiven_project.kafka-mm-project1.project
  service_name = aiven_service.target.service_name
  topic_name = "topic-b"
  partitions = 3
  replication = 2
}