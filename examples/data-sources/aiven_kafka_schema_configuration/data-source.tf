data "aiven_kafka_schema_configuration" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  /* COMPUTED FIELDS
  compatibility_level = "FULL"
  */
}
