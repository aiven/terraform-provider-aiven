data "aiven_kafka" "example_kafka" {
  project      = data.aiven_project.example_project.project
  service_name = "example-kafka"
}
