data "aiven_organization_billing_group_list" "example" {
  organization_id = "org1a23f456789"

  /* COMPUTED FIELDS
  billing_groups {
    billing_address_id  = "foo"
    billing_group_id    = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    organization_id     = "org1a23f456789"
    shipping_address_id = "foo"
    vat_id              = "foo"
    billing_contact_emails {
      email = "foo@example.com"
    }
    billing_emails {
      email = "foo@example.com"
    }
    billing_group_name  = "foo"
    custom_invoice_text = "foo"
    payment_method {
      payment_method_id   = "foo"
      payment_method_type = "aws_subscription"
    }
  }
  */
}
