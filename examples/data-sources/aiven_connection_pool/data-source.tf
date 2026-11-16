data "aiven_connection_pool" "example" {
  project      = "my-project"
  service_name = "my-pg"
  pool_name    = "mypool-x-y-z"

  /* COMPUTED FIELDS
  connection_uri = "foo"
  database_name  = "testdb"
  pool_mode      = "transaction"
  pool_size      = 10
  username       = "testuser"
  */
}
