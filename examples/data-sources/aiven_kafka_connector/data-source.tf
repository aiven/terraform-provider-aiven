data "aiven_kafka_connector" "kafka-os-connector" {
  project        = data.aiven_project.example_project.project
  service_name   = aiven_kafka.example_kafka.service_name
  connector_name = "kafka-opensearch-connector"
}
