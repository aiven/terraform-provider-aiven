data "aiven_organization_address" "example" {
  organization_id = "org1a23f456789"
  address_id      = "foo"

  /* COMPUTED FIELDS
  address_lines = ["Main Street 1"]
  city          = "Helsinki"
  country_code  = "FI"
  create_time   = "2021-01-01T00:00:00Z"
  name          = "Aiven Oy"
  state         = "foo"
  update_time   = "2021-01-01T00:00:00Z"
  zip_code      = "01234"
  */
}
