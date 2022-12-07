data "aiven_kafka_topic" "mytesttopic" {
  project      = aiven_project.myproject.project
  service_name = aiven_kafka.myservice.service_name
  topic_name   = "<TOPIC_NAME>"
}

