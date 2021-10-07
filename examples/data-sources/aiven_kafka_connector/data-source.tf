data "aiven_kafka_connector" "kafka-es-con1" {
    project = aiven_project.kafka-con-project1.project
    service_name = aiven_service.kafka-service1.service_name
    connector_name = "kafka-es-con1"
}

