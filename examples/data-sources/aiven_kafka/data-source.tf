data "aiven_kafka" "kafka1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-kafka1"
}

