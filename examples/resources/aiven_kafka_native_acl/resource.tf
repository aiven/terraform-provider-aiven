resource "aiven_kafka_native_acl" "example" {
  project         = "my-project" // Force new
  service_name    = "my-kafka" // Force new
  operation       = "Read" // Force new
  pattern_type    = "LITERAL" // Force new
  permission_type = "ALLOW" // Force new
  principal       = "User:alice" // Force new
  resource_name   = "consumer-group-1." // Force new
  resource_type   = "Topic" // Force new

  // OPTIONAL FIELDS
  host = "*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
