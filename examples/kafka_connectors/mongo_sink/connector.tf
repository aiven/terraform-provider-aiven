data "aiven_service_component" "schema_registry" {
  project      = aiven_kafka.kafka-service1.project
  service_name = aiven_kafka.kafka-service1.service_name
  component    = "schema_registry"
  route        = "dynamic"

  depends_on = [
    aiven_kafka.kafka-service1
  ]
}

locals {
  schema_registry_uri = "https://${data.aiven_service_user.kafka_admin.username}:${data.aiven_service_user.kafka_admin.password}@${data.aiven_service_component.schema_registry.host}:${data.aiven_service_component.schema_registry.port}"
}

# Kafka Mongo Sink connector
resource "aiven_kafka_connector" "kafka-mongo-sink-con1" {
  project        = aiven_project.kafka-con-project1.project
  service_name   = aiven_kafka.kafka-service1.service_name
  connector_name = "mongo-sink"

  config = {
    "name"            = "mongo-sink"
    "connector.class" = "com.mongodb.kafka.connect.MongoSinkConnector"
    "topics"          = "test-kafka-topic1"
    "tasks.max"       = 1

    // MongoDB connect settings
    "connection.uri" = var.mongo_uri
    "database"       = "test-mongo-db"
    "max.batch.size" = 1

    // Common Settings
    "key.converter"                               = "io.confluent.connect.avro.AvroConverter"
    "key.converter.schema.registry.url"           = local.schema_registry_uri
    "key.converter.basic.auth.credentials.source" = "URL"
    "key.converter.schemas.enable"                = "true"

    "value.converter" : "io.confluent.connect.avro.AvroConverter",
    "value.converter.schema.registry.url" : local.schema_registry_uri,
    "value.converter.basic.auth.credentials.source" : "URL",
    "value.converter.schemas.enable" : "true",
  }
}