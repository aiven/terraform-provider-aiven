data "aiven_kafka_topic" "example_topic" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_kafka.example_kafka.service_name
  topic_name   = "example-topic"
}
