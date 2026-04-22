resource "aiven_organization_address" "example" {
  organization_id = "org1a23f456789" // Force new
  address_lines   = ["Main Street 1"]
  city            = "Helsinki"
  country_code    = "FI"
  name            = "Aiven Oy"

  // OPTIONAL FIELDS
  state    = "foo"
  zip_code = "01234"

  /* COMPUTED FIELDS
  address_id  = "foo"
  create_time = "2021-01-01T00:00:00Z"
  update_time = "2021-01-01T00:00:00Z"
  */
}
