resource "aiven_kafka_acl" "example_acl" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_kafka.example_kafka.service_name
  topic        = "example-topic"
  permission   = "admin"
  username     = "example-user"
}
