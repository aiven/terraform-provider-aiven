data "aiven_billing_group" "example" {
  billing_group_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  card_id                 = "9330c086-8781-11e5-89ff-5404a64abfef"
  parent_id               = "foo"
  vat_id                  = "FI27957435"
  address_lines           = ["Main Street 1"]
  billing_contact_emails  = ["jane@example.com"]
  billing_currency        = "USD"
  billing_emails          = ["test@example.com"]
  billing_extra_text      = "Purchase order: PO100018"
  city                    = "Helsinki"
  company                 = "My Company"
  copy_from_billing_group = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  country_code            = "FI"
  name                    = "my billing group"
  state                   = "foo"
  zip_code                = "01234"
  */
}
