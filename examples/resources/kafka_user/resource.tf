resource "aiven_kafka_user" "foo" {
  service_name = aiven_kafka.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}