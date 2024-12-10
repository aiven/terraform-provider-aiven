resource "aiven_kafka_native_acl" "example_acl" {
  project         = data.aiven_project.example_project.project
  service_name    = aiven_kafka.example_kafka.service_name
  resource_type   = "Topic"
  resource_name   = "example-topic"
  principal       = "User:example-user"
  operation      = "Read"
  pattern_type    = "LITERAL"
  permission_type = "ALLOW"
  host            = "198.51.100.0"
}
