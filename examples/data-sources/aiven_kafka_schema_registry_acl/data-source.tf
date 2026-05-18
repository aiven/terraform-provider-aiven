data "aiven_kafka_schema_registry_acl" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  // LOOKUP — provide `acl_id`, or all of: `permission`, `resource` and `username`
  acl_id        = "foo"
  // permission = "schema_registry_read"
  // resource   = "Config:"
  // username   = "admin*"
}
