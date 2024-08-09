resource "aiven_billing_group" "example_billing_group" {
  name             = "example-billing-group"
  billing_currency = "USD"
  vat_id           = "123ABC"
  parent_id        =  data.aiven_organization.main.id
}

resource "aiven_project" "example_project" {
  project       = "example-project"
  billing_group = aiven_billing_group.example_billing_group.id
}
