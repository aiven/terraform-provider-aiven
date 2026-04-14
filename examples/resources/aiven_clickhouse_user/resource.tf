resource "aiven_clickhouse_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-clickhouse" // Force new
  username     = "alice" // Force new

  // OPTIONAL FIELDS
  password_wo         = "password123"
  password_wo_version = 42

  /* COMPUTED FIELDS
  uuid     = "foo"
  required = true
  */
}
