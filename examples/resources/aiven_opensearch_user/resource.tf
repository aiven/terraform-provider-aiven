resource "aiven_opensearch_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-opensearch" // Force new
  username     = "testuser" // Force new

  // OPTIONAL FIELDS
  password_wo         = "password123"
  password_wo_version = 42

  /* COMPUTED FIELDS
  type = "foo"
  */
}
