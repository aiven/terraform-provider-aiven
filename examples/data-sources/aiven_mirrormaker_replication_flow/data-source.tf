data "aiven_mirrormaker_replication_flow" "f1" {
  project        = aiven_project.kafka-mm-project1.project
  service_name   = aiven_service.mm.service_name
  source_cluster = aiven_service.source.service_name
  target_cluster = aiven_service.target.service_name
}

