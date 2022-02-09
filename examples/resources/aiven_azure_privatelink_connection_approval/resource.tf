resource "aiven_static_ip" "static_ips" {
  count      = 2
  project    = var.project_id
  cloud_name = var.region
}

resource "aiven_pg" "default" {
  service_name   = "postgres"
  project        = var.aiven_project_id
  project_vpc_id = var.aiven_project_vpc_id
  cloud_name     = var.region
  plan           = var.plan
  static_ips     = [for sip in aiven_static_ip.static_ips : sip.static_ip_address_id]

  pg_user_config {
    pg_version = 13
    static_ips = true
    privatelink_access {
      pg        = true
      pgbouncer = true
    }
  }

}

resource "aiven_azure_privatelink" "privatelink" {
  project      = var.aiven_project_id
  service_name = aiven_pg.default.name
  user_subscription_ids = [
    var.azure_subscription_id
  ]
}

resource "azurerm_private_endpoint" "endpoint" {
  name                = "postgres-endpoint"
  location            = var.region
  resource_group_name = var.azure_resource_group.name
  subnet_id           = var.azure_subnet_id
  private_service_connection {
    name                           = aiven_pg.default.name
    private_connection_resource_id = aiven_azure_privatelink.privatelink.azure_service_id
    is_manual_connection           = true
    request_message                = aiven_pg.default.name
  }
  depends_on = [
    aiven_azure_privatelink.privatelink,
  ]
}

resource "aiven_azure_privatelink_connection_approval" "approval" {
  project             = var.aiven_project_id
  service_name        = aiven_pg.default
  endpoint_ip_address = azurerm_private_endpoint.endpoint.private_service_connection[0].private_ip_address
}
