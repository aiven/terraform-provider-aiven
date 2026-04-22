resource "aiven_clickhouse_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-clickhouse" // Force new
  username     = "alice" // Force new

  // OPTIONAL FIELDS
  password_wo         = "password123"
  password_wo_version = 1

  /* COMPUTED FIELDS
  uuid     = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  required = true
  */
}
