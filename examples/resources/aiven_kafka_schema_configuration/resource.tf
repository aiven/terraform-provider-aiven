resource "aiven_kafka_schema_configuration" "example" {
  project             = aiven_kafka.example_kafka.project
  service_name        = aiven_kafka.example_kafka.service_name
  compatibility_level = "FULL"
}

resource "aiven_kafka" "example_kafka" {
  project      = "my-project"
  plan         = "business-4"
  service_name = "example-kafka"

  kafka_user_config {
    schema_registry = true
  }
}
