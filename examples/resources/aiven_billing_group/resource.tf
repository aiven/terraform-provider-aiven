resource "aiven_billing_group" "bybg1" {
  name             = "bybg1"
  billing_currency = "USD"
  vat_id           = "123ABC"
}

resource "aiven_project" "pr1" {
  project       = "pr1"
  billing_group = aiven_billing_group.bybg1.id
}
