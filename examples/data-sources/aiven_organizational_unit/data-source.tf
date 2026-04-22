data "aiven_organizational_unit" "example" {
  // REQUIRED EXACTLY ONE
  id      = "foo"
  // name = "Aiven Ltd"

  /* COMPUTED FIELDS
  parent_id   = "a22ba494e096"
  tenant_id   = "foo"
  create_time = "2021-01-01T00:00:00Z"
  update_time = "2021-01-01T00:00:00Z"
  */
}
