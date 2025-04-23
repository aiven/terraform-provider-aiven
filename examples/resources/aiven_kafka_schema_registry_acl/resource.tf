resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = aiven_project.kafka-schemas-project1.project
  service_name = aiven_kafka.kafka-service1.service_name
  resource     = "Subject:topic-1"
  username     = "group-user-*"
  permission   = "schema_registry_read"
}
