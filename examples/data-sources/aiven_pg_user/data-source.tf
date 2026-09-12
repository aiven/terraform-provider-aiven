data "aiven_pg_user" "example" {
  project      = "my-project"
  service_name = "my-pg"
  username     = "testuser"

  /* COMPUTED FIELDS
  access_cert              = "foo"
  access_key               = "foo"
  mysql_grants             = ["SELECT", "DELETE"]
  password                 = "password123"
  password_encryption_type = "md5"
  pg_allow_replication     = true
  type                     = "foo"
  */
}
