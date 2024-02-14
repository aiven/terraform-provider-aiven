# ClickHouse service in the same region
resource "aiven_clickhouse" "clickhouse" {
  project                 = aiven_project.clickhouse_kafka_source.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-beta-16"
  service_name            = "clickhouse-gcp-eu"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

# ClickHouse service integration with a Kafka source topic containing
# edge measurements in JSON format.
# This will create a `service_kafka-gcp-eu` database with a
# `edge_measurements_raw` using the Kafka ClickHouse Engine.
resource "aiven_service_integration" "clickhouse_kafka_source" {
  project                  = aiven_project.clickhouse_kafka_source.project
  integration_type         = "clickhouse_kafka"
  source_service_name      = aiven_kafka.kafka.service_name
  destination_service_name = aiven_clickhouse.clickhouse.service_name
  clickhouse_kafka_user_config {
    tables {
      name        = "edge_measurements_raw"
      group_name  = "clickhouse-ingestion"
      data_format = "JSONEachRow"
      columns {
        name = "sensor_id"
        type = "UInt64"
      }
      columns {
        name = "ts"
        type = "DateTime64(6)"
      }
      columns {
        name = "key"
        type = "LowCardinality(String)"
      }
      columns {
        name = "value"
        type = "Float64"
      }
      topics {
        name = aiven_kafka_topic.edge_measurements.topic_name
      }
    }
  }
}

# ClickHouse database that can be used to run analytics over the ingested data
resource "aiven_clickhouse_database" "iot_analytics" {
  project      = aiven_project.clickhouse_kafka_source.project
  service_name = aiven_clickhouse.clickhouse.service_name
  name         = "iot_analytics"
}
