resource "aiven_kafka" "kafka" {
  project                 = var.project_name
  cloud_name              = "google-europe-north1"
  plan                    = "business-4"
  service_name            = var.kafka_service_name
  maintenance_window_dow  = "saturday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka_rest      = true
    kafka_connect   = false
    schema_registry = true

    kafka {
      auto_create_topics_enable  = true
      num_partitions             = 3
      default_replication_factor = 2
      min_insync_replicas        = 2
    }

    kafka_authentication_methods {
      certificate = true
    }

  }
}

resource "aiven_kafka_connect" "connect" {
  project                 = var.project_name
  cloud_name              = "google-europe-north1"
  plan                    = "business-4"
  service_name            = var.kafka_connect_service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = false
    }
  }
}

resource "aiven_service_integration" "kafka_and_connect_integration" {
  project                  = var.project_name
  integration_type         = "kafka_connect"
  source_service_name      = aiven_kafka.kafka.service_name
  destination_service_name = aiven_kafka_connect.connect.service_name

  kafka_connect_user_config {
    kafka_connect {
      group_id             = "connect"
      status_storage_topic = "__connect_status"
      offset_storage_topic = "__connect_offsets"
    }
  }
}
