data "aiven_kafka_acl" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  // LOOKUP — provide `acl_id`, or all of: `permission`, `topic` and `username`
  acl_id        = "foo"
  // permission = "readwrite"
  // topic      = "top*"
  // username   = "admin*"
}
