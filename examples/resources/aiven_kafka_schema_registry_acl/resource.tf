resource "aiven_kafka_schema_registry_acl" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka" // Force new
  permission   = "schema_registry_read" // Force new
  resource     = "Config:" // Force new
  username     = "admin*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
