resource "aiven_byoc_azure_provision" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  azure_subscription_id       = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  azure_tenant_id             = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  aiven_azure_cce_client_id = "foo"
  custom_cloud_names        = ["foo"]
  state                     = "active"
  */
}
