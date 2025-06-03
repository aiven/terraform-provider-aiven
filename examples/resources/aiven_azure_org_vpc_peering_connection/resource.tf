resource "aiven_organization_vpc" "example_vpc" {
  organization_id = data.aiven_organization.example.id
  cloud_name      = "azure-germany-westcentral"
  network_cidr    = "10.0.0.0/24"
}

resource "aiven_azure_org_vpc_peering_connection" "example_peering" {
  organization_id       = aiven_organization_vpc.example_vpc.organization_id
  organization_vpc_id   = aiven_organization_vpc.example_vpc.organization_vpc_id
  azure_subscription_id = "12345678-1234-1234-1234-123456789012"
  vnet_name             = "my-vnet"
  peer_resource_group   = "my-resource-group"
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321"
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555"
}
