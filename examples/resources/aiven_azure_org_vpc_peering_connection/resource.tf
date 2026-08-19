resource "aiven_azure_org_vpc_peering_connection" "example" {
  organization_id       = "org1a23f456789" // Force new
  organization_vpc_id   = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  azure_subscription_id = "12345678-1234-1234-1234-123456789012" // Force new
  vnet_name             = "my-vnet" // Force new
  peer_resource_group   = "my-resource-group" // Force new
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321" // Force new
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555" // Force new

  /* COMPUTED FIELDS
  peering_connection_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  */
}
