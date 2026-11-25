resource "aiven_kafka_schema" "example" {
  project             = aiven_kafka.example_kafka.project
  service_name        = aiven_kafka.example_kafka.service_name
  subject_name        = "my-topic-value"
  compatibility_level = "FULL"
  schema = jsonencode({
    fields = [{
      name = "id"
      type = "string"
    }]
    name = "Order"
    type = "record"
  })
}

resource "aiven_kafka" "example_kafka" {
  project      = "my-project"
  plan         = "business-4"
  service_name = "example-kafka"

  kafka_user_config {
    schema_registry = true
  }
}
