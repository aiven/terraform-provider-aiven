resource "aiven_kafka" "kafka" {
  project                 = var.avn_project
  service_name            = var.kafka_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    // Enables Kafka Connectors
    kafka_connect = true
    kafka_version = "3.2"

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "kafka-topic" {
  project      = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  topic_name   = var.kafka_topic_name
  partitions   = 3
  replication  = 2
}

resource "aiven_opensearch" "os" {
  project                 = var.avn_project
  service_name            = var.os_name
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_kafka_connector" "kafka-os-connector" {
  project        = aiven_kafka.kafka.project
  service_name   = aiven_kafka.kafka.service_name
  connector_name = var.kafka_connector_name

  config = {
    "topics"                         = aiven_kafka_topic.kafka-topic.topic_name
    "connector.class"                = "io.aiven.kafka.connect.opensearch.OpensearchSinkConnector"
    "type.name"                      = "os-connector"
    "name"                           = var.kafka_connector_name
    "connection.url"                 = "https://${aiven_opensearch.os.service_host}:${aiven_opensearch.os.service_port}"
    "connection.username"            = aiven_opensearch.os.service_username
    "connection.password"            = aiven_opensearch.os.service_password
    "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
    "tasks.max"                      = 1
    "schema.ignore"                  = true
    "value.converter.schemas.enable" = false
  }
}
