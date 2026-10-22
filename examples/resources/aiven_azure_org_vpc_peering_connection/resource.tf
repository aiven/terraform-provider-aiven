resource "aiven_azure_org_vpc_peering_connection" "example" {
  organization_id       = "org1a23f456789" // Force new
  organization_vpc_id   = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  azure_subscription_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  vnet_name             = "my-vnet" // Force new
  peer_resource_group   = "my-resource-group" // Force new
  peer_azure_app_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  peer_azure_tenant_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new

  /* COMPUTED FIELDS
  peering_connection_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  */
}
