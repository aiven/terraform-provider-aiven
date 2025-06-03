terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">= 3.8.0, < 5.0.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "=2.30.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=3.30.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_token
}

provider "azuread" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = var.azure_client_secret
  tenant_id     = "00000000-0000-0000-0000-000000000000"
}

provider "azurerm" {
  features {}
  subscription_id = "00000000-0000-0000-0000-000000000000"
  client_id       = "00000000-0000-0000-0000-000000000000"
  client_secret   = var.azure_client_secret
  tenant_id       = "00000000-0000-0000-0000-000000000000"
}

# Your Aiven project
data "aiven_project" "example_project" {
  project = var.aiven_project_name
}

# Create a VPC for your Aiven project
resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "azure-germany-westcentral"
  network_cidr = "192.168.1.0/24"

  timeouts {
    create = "15m"
  }
}

# Create an Azure resource group
resource "azurerm_resource_group" "azure_resource_group" {
  location = "germanywestcentral"
  name     = "example-azure-resource-group"
}

# Create a virtual network within the Azure resource group
resource "azurerm_virtual_network" "vnet" {
  name                = "example-azure-virtual-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.azure_resource_group.location
  resource_group_name = azurerm_resource_group.azure_resource_group.name
}

# Create an Azure application object
resource "azuread_application" "example_app" {
  display_name     = "example-azure-application"
  sign_in_audience = "AzureADandPersonalMicrosoftAccount"

  api {
    requested_access_token_version = 2
  }
}


# Create a service principal for the application object
resource "azuread_service_principal" "app_principal" {
  client_id = azuread_application.example_app.client_id
}

# Set a password for the application object.
resource "azuread_application_password" "app_password" {
  application_id = azuread_application.example_app.application_id
}

# Grant the service principal permissions to peer.
resource "azurerm_role_assignment" "app_role" {
  role_definition_name = "Network Contributor"
  principal_id         = azuread_service_principal.app_principal.object_id
  scope                = azurerm_virtual_network.vnet.id
}

# Create a service principal for the Aiven application object with the application_id hardcoded
resource "azuread_service_principal" "aiven_app_principal" {
  client_id = "55f300d4-fc50-4c5e-9222-e90a6e2187fb"
}

# Create a custom role for the Aiven application object
data "azurerm_subscription" "example_subscription" {
  subscription_id = "00000000-0000-0000-0000-000000000000"
}

resource "azurerm_role_definition" "example_role" {
  name        = "example-azure-role-definition"
  description = "Allows creating a peering to, but not from, VNnets in scope."
  scope       = "/subscriptions/${data.azurerm_subscription.example_subscription.subscription_id}"

  permissions {
    actions = ["Microsoft.Network/virtualNetworks/peer/action"]
  }

  assignable_scopes = [
    "/subscriptions/${data.azurerm_subscription.example_subscription.subscription_id}"
  ]
}

# Assign the custom role to the Aiven service principal
resource "azurerm_role_assignment" "aiven_role_assignment" {
  role_definition_id = azurerm_role_definition.example_role.role_definition_resource_id
  principal_id       = azuread_service_principal.aiven_app_principal.object_id
  scope              = azurerm_virtual_network.vnet.id

  depends_on = [
    azuread_service_principal.aiven_app_principal,
    azurerm_role_assignment.app_role
  ]
}

# Create a peering connection from the Aiven project VPC
resource "aiven_azure_vpc_peering_connection" "peering_connection" {
  vpc_id                = aiven_project_vpc.aiven_vpc.id
  peer_resource_group   = azurerm_resource_group.azure_resource_group.name
  azure_subscription_id = data.azurerm_subscription.example_subscription.subscription_id
  vnet_name             = azurerm_virtual_network.vnet.name
  peer_azure_app_id     = azuread_application.example_app.application_id
  peer_azure_tenant_id  = "00000000-0000-0000-0000-000000000000"

  depends_on = [
    azurerm_role_assignment.aiven_role_assignment
  ]
}

# Create peering connection from the VNet to the Aiven project VPC's VNet
provider "azurerm" {
  features {}
  alias                = "app"
  client_id            = azuread_application.example_app.application_id
  client_secret        = azuread_application_password.app_password.value
  subscription_id      = data.azurerm_subscription.example_subscription.subscription_id
  tenant_id            = "00000000-0000-0000-0000-000000000000"
  auxiliary_tenant_ids = [azuread_service_principal.aiven_app_principal.application_tenant_id]
}

resource "azurerm_virtual_network_peering" "network_peering" {
  provider                     = azurerm.app
  name                         = "example-azure-virtual-network-peering"
  remote_virtual_network_id    = aiven_azure_vpc_peering_connection.peering_connection.state_info["to-network-id"]
  resource_group_name          = azurerm_resource_group.azure_resource_group.name
  virtual_network_name         = azurerm_virtual_network.vnet.name
  allow_virtual_network_access = true

  depends_on = [
    aiven_azure_vpc_peering_connection.peering_connection
  ]
}
