data "aiven_azure_vpc_peering_connection" "foo" {
  vpc_id                = data.aiven_project_vpc.vpc.id
  azure_subscription_id = "xxxxxx"
  peer_resource_group   = "my-pr1"
  vnet_name             = "my-vnet1"
  peer_azure_app_id     = "xxxxxx"
  peer_azure_tenant_id  = "xxxxxx"
}
