data "aiven_clickhouse_user" "example" {
  project      = "my-project"
  service_name = "my-clickhouse"

  // REQUIRED EXACTLY ONE
  uuid     = "foo"
  username = "alice"

  /* COMPUTED FIELDS
  password = "!@$password12345"
  required = true
  */
}
