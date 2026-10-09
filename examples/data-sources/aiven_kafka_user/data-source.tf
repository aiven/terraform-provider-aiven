data "aiven_kafka_user" "example" {
  project      = "my-project"
  service_name = "my-kafka"
  username     = "testuser"

  /* COMPUTED FIELDS
  access_cert              = "foo"
  access_key               = "foo"
  password                 = "password123"
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
