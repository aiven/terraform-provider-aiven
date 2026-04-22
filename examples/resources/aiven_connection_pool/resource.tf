resource "aiven_connection_pool" "example" {
  project       = "my-project" // Force new
  service_name  = "foo" // Force new
  pool_name     = "mypool-x-y-z" // Force new
  database_name = "testdb" // Force new

  // OPTIONAL FIELDS
  pool_mode = "transaction"
  pool_size = 10
  username  = "testuser"

  /* COMPUTED FIELDS
  connection_uri = "foo"
  */
}
