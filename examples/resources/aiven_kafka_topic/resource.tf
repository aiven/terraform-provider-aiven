resource "aiven_kafka_topic" "example_topic" {
  project                = data.aiven_project.example_project.project
  service_name           = aiven_kafka.example_kafka.service_name
  topic_name             = "example-topic"
  partitions             = 5
  replication            = 3
  termination_protection = true

  config {
    flush_ms                       = 10
    cleanup_policy                 = "compact,delete"
  }


  timeouts {
    create = "1m"
    read   = "5m"
  }
}