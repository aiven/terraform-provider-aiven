# Kafka Elasticsearch sink connector
resource "aiven_kafka_connector" "kafka-es-con1" {
  project = aiven_project.kafka-con-project1.project
  service_name = aiven_kafka.kafka-service1.service_name
  connector_name = "kafka-es-con1"

  config = {
    "topics" = aiven_kafka_topic.kafka-topic1.topic_name
    "connector.class" : "io.aiven.connect.elasticsearch.ElasticsearchSinkConnector"
    "type.name" = "es-connector"
    "name" = "kafka-es-con1"
    "connection.url" = aiven_elasticsearch.es-service1.service_uri
  }
}