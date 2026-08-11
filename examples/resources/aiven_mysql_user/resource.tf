resource "aiven_mysql_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-mysql" // Force new
  username     = "testuser" // Force new

  // OPTIONAL FIELDS
  password_wo         = "password123"
  password_wo_version = 1
  authentication      = "caching_sha2_password"

  /* COMPUTED FIELDS
  access_cert              = "foo"
  access_key               = "foo"
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
