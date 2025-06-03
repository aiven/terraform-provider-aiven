resource "aiven_kafka_quota" "example_quota" {
  project            = data.aiven_project.foo.project
  service_name       = aiven_kafka.example_kafka.service_name
  user               = "example-kafka-user"
  client_id          = "example_client"
  consumer_byte_rate = 1000
  producer_byte_rate = 1000
  request_percentage = 50
}
