data "aiven_mysql_user" "example" {
  project      = "my-project"
  service_name = "my-mysql"
  username     = "testuser"

  /* COMPUTED FIELDS
  access_cert    = "foo"
  access_key     = "foo"
  authentication = "caching_sha2_password"
  password       = "password123"
  type           = "foo"
  */
}
