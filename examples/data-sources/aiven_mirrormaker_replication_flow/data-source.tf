data "aiven_mirrormaker_replication_flow" "example_replication_flow" {
  project        = data.aiven_project.example_project.project
  service_name   = aiven_kafka.example_kafka.service_name
  source_cluster = aiven_kafka.source.service_name
  target_cluster = aiven_kafka.target.service_name
}

