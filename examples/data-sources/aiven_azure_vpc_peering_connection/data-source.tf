data "aiven_azure_vpc_peering_connection" "example" {
  azure_subscription_id = "12345678-1234-1234-1234-123456789012"
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321"
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555"
  vpc_id                = "example-project/example-vpc"
  peer_resource_group   = "my-resource-group"
  vnet_name             = "my-vnet"

  /* COMPUTED FIELDS
  id                    = "example-project/example-vpc/12345678-1234-1234-1234-123456789012/my-vnet"
  peering_connection_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
