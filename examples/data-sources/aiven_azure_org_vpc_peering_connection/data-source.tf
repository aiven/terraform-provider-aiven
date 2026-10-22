data "aiven_azure_org_vpc_peering_connection" "example" {
  organization_id       = "org1a23f456789"
  organization_vpc_id   = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  azure_subscription_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  vnet_name             = "my-vnet"
  peer_resource_group   = "my-resource-group"

  /* COMPUTED FIELDS
  peer_azure_app_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peer_azure_tenant_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peering_connection_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  */
}
