resource "aiven_azure_privatelink" "example" {
  project               = "my-project" // Force new
  service_name          = "foo" // Force new
  user_subscription_ids = ["adcf7194-d877-4505-a47a-91fefd96e3b8"]

  /* COMPUTED FIELDS
  azure_service_id    = "foo"
  azure_service_alias = "foo"
  state               = "active"
  */
}
