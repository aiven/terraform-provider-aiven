resource "aiven_pg_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-pg" // Force new
  username     = "testuser" // Force new

  // OPTIONAL FIELDS
  password_wo          = "password123"
  password_wo_version  = 1
  pg_allow_replication = true

  /* COMPUTED FIELDS
  access_cert = "foo"
  access_key  = "foo"
  type        = "foo"
  */
}
