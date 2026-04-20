data "aiven_clickhouse_user" "example" {
  project      = "my-project"
  service_name = "my-clickhouse"

  // REQUIRED EXACTLY ONE
  uuid        = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  // username = "alice"

  /* COMPUTED FIELDS
  password = "!@$password12345"
  required = true
  */
}
