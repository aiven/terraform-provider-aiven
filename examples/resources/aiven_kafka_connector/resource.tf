resource "aiven_kafka_connector" "kafka-os-con1" {
  project        = aiven_project.kafka-con-project1.project
  service_name   = aiven_kafka.kafka-service1.service_name
  connector_name = "kafka-os-con1"

  config = {
    "topics" = aiven_kafka_topic.kafka-topic1.topic_name
    "connector.class" : "io.aiven.kafka.connect.opensearch.OpensearchSinkConnector"
    "type.name"      = "os-connector"
    "name"           = "kafka-os-con1"
    "connection.url" = "${aiven_opensearch.os-service1.service_host}:${aiven_opensearch.os-service1.service_port}"
    "connection.username" = aiven_opensearch.os-service1.service_username
    "connection.password" = aiven_opensearch.os-service1.service_password
  }
}
