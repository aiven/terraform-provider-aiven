resource "aiven_mirrormaker_replication_flow" "f1" {
  project        = aiven_project.kafka-mm-project1.project
  service_name   = aiven_kafka.mm.service_name
  source_cluster = aiven_kafka.source.service_name
  target_cluster = aiven_kafka.target.service_name
  enable         = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]
}
