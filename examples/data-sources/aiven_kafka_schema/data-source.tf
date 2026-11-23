data "aiven_kafka_schema" "example" {
  project      = "my-project"
  service_name = "my-kafka"
  subject_name = "my-topic-value"

  /* COMPUTED FIELDS
  compatibility_level = "FULL"
  references {
    name    = "container1.proto"
    subject = "topic1-values"
    version = 42
  }
  schema = jsonencode({
    fields = [{
      name = "id"
      type = "string"
    }]
    name = "Order"
    type = "record"
  })
  schema_type = "AVRO"
  version     = 42
  */
}
