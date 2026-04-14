data "aiven_organization_payment_method_list" "example" {
  organization_id = "org1a23f456789"

  /* COMPUTED FIELDS
  payment_methods {
    payment_method_id   = "foo"
    payment_method_type = "aws_subscription"
  }
  */
}
