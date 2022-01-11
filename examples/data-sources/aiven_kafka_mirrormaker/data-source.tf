data "aiven_kafka_mirrormaker" "mm1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-mm1"
}

