data "aiven_organization_billing_group" "example" {
  organization_id  = "org1a23f456789"
  billing_group_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  billing_address_id  = "addr4b1ff1ceeaa"
  shipping_address_id = "addr4b1ff1ceeaa"
  vat_id              = "FI12345678"
  billing_contact_emails {
    email = "jane@example.com"
  }
  billing_emails {
    email = "jane@example.com"
  }
  billing_group_name  = "Default billing group for the organization"
  custom_invoice_text = "Extra billing text"
  payment_method {
    payment_method_id   = "pm4b1ff1ceeaa"
    payment_method_type = "credit_card"
  }
  */
}
