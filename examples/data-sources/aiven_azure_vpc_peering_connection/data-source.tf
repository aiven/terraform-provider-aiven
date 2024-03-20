data "aiven_azure_vpc_peering_connection" "azure_to_aiven_peering" {
  vpc_id                = data.aiven_project_vpc.example_vpc.id
  azure_subscription_id = "00000000-0000-0000-0000-000000000000"
  peer_resource_group   = "example-resource-group"
  vnet_name             = "example-vnet"
  peer_azure_app_id     = "00000000-0000-0000-0000-000000000000"
  peer_azure_tenant_id  = "00000000-0000-0000-0000-000000000000"
}
