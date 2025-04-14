resource "aiven_kafka_connector" "kafka-os-connector" {
  project        = data.aiven_project.example_project.project
  service_name   = aiven_kafka.example_kafka.service_name
  connector_name = "kafka-opensearch-connector"

  config = {
    "name"                = "kafka-opensearch-connector" # Must be the same as the connector_name.
    "topics"              = aiven_kafka_topic.example_topic.topic_name
    "connector.class"     = "io.aiven.kafka.connect.opensearch.OpensearchSinkConnector"
    "type.name"           = "os-connector"
    "connection.url"      = aiven_opensearch.example_os.service_uri
    "connection.username" = aiven_opensearch.example_os.service_username
    "connection.password" = aiven_opensearch.example_os.service_password
  }
}
