data "aiven_kafka_connect" "example_kafka_connect" {
  project      = data.aiven_project.example_project.project
  service_name = "example-connect-service"
}
