data "aiven_kafka_user" "user" {
  service_name = "my-service"
  project      = "my-project"
  username     = "user1"
}