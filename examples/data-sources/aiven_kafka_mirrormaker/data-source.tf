data "aiven_kafka_mirrormaker" "example_mirrormaker" {
  project      = data.aiven_project.example_project.project
  service_name = "example-mirrormaker-service"
}