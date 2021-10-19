resource "aiven_kafka_schema_configuration" "config" {
    project = aiven_project.kafka-schemas-project1.project
    service_name = aiven_kafka.kafka-service1.service_name
    compatibility_level = "BACKWARD"
}
