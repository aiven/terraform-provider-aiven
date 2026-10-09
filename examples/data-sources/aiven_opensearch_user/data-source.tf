data "aiven_opensearch_user" "example" {
  project      = "my-project"
  service_name = "my-opensearch"
  username     = "testuser"

  /* COMPUTED FIELDS
  password                 = "password123"
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
