resource "aiven_organizational_unit" "example" {
  parent_id = "a22ba494e096"
  name      = "Aiven Ltd"

  /* COMPUTED FIELDS
  id          = "foo"
  tenant_id   = "foo"
  create_time = "2021-01-01T00:00:00Z"
  update_time = "2021-01-01T00:00:00Z"
  */
}
