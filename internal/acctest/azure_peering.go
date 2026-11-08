package acctest

import "fmt"

// AzurePeeringConfig is shared by project and organization VPC peering tests.
type AzurePeeringConfig struct {
	Project        string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	AivenAppID     string `envconfig:"AIVEN_AZURE_APP_ID" required:"true"`
	TenantID       string `envconfig:"AZURE_TENANT_ID" required:"true"`
	SubscriptionID string `envconfig:"AZURE_SUBSCRIPTION_ID" required:"true"`
}

// AzurePeeringFixture creates a peer VNet and grants Aiven permission to peer with it.
// The existing Aiven service principal is read-only and is never owned by the test.
func AzurePeeringFixture(prefix string, s *AzurePeeringConfig) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
  tenant_id       = %[2]q
  subscription_id = %[3]q
}

provider "azuread" {
  tenant_id = %[2]q
}

data "azuread_client_config" "current" {}

data "azurerm_subscription" "subscription" {
  subscription_id = %[3]q
}

resource "azurerm_resource_group" "resource_group" {
  location = "germanywestcentral"
  name     = "%[1]s-resource-group"
}

resource "azurerm_virtual_network" "virtual_network" {
  name                = "%[1]s-virtual-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azuread_application" "application" {
  display_name = "%[1]s-application"
  # Keep the caller as an owner for credential management and cleanup.
  owners = [data.azuread_client_config.current.object_id]
  # Accounts in any organizational directory (multi-tenant)
  #  https://learn.microsoft.com/en-us/azure/active-directory/develop/supported-accounts-validation
  sign_in_audience = "AzureADandPersonalMicrosoftAccount"

  api {
    requested_access_token_version = 2
  }
}

resource "azuread_service_principal" "app_principal" {
  application_id = azuread_application.application.application_id
  owners         = [data.azuread_client_config.current.object_id]
}

resource "azuread_application_password" "app_password" {
  application_object_id = azuread_application.application.object_id
}

resource "azurerm_role_assignment" "app_role_assignment" {
  role_definition_name = "Network Contributor"
  principal_id         = azuread_service_principal.app_principal.object_id
  scope                = azurerm_virtual_network.virtual_network.id
}

# The Aiven service principal is shared across peerings and must survive cleanup.
data "azuread_service_principal" "aiven_principal" {
  application_id = %[4]q
}

resource "azurerm_role_definition" "role_definition" {
  name        = "%[1]s-role-definition"
  description = "Allows creating a peering to vnets in scope (but not from)"
  scope       = "/subscriptions/${data.azurerm_subscription.subscription.subscription_id}"

  permissions {
    actions = ["Microsoft.Network/virtualNetworks/peer/action"]
  }

  assignable_scopes = [
    "/subscriptions/${data.azurerm_subscription.subscription.subscription_id}"
  ]
}

resource "azurerm_role_assignment" "aiven_role_assignment" {
  role_definition_id = azurerm_role_definition.role_definition.role_definition_resource_id
  principal_id       = data.azuread_service_principal.aiven_principal.object_id
  scope              = azurerm_virtual_network.virtual_network.id

  depends_on = [
    azurerm_role_assignment.app_role_assignment
  ]
}

`, prefix, s.TenantID, s.SubscriptionID, s.AivenAppID)
}
