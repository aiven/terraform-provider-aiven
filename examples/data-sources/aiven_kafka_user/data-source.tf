data "aiven_kafka_user" "example" {
  project      = "my-project"
  service_name = "my-kafka"
  username     = "testuser"

  /* COMPUTED FIELDS
  access_cert = "foo"
  access_key  = "foo"
  password    = "password123"
  type        = "foo"
  */
}
