resource "aiven_azure_vpc_peering_connection" "example" {
  azure_subscription_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  peer_azure_app_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  peer_azure_tenant_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  peer_resource_group   = "my-resource-group" // Force new
  vnet_name             = "my-vnet" // Force new

  // OPTIONAL FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  project        = "example-project"

  /* COMPUTED FIELDS
  id                    = "example-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d/2a3b4c5d-6e7f-4a8b-9c0d-1e2f3a4b5c6d/my-vnet"
  peering_connection_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  create_time           = "2021-01-01T00:00:00Z"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
