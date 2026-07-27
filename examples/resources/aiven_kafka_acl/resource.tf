resource "aiven_kafka_acl" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka" // Force new
  permission   = "readwrite" // Force new
  topic        = "top*" // Force new
  username     = "admin*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
