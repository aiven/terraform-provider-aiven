package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenAzureVPCPeeringConnection_basic(t *testing.T) {
	var s azureSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	// This test requires dynamic provider configuration (user secret made in Azure).
	// Terraform doesn't support dynamic configuration https://github.com/hashicorp/terraform/issues/25244
	// We run two configs here on by one.

	prefix := "test-tf-acc-vpcpeering-" + acctest.RandString(7)
	configOne := testAccVPCPeeringConnectionAzureResourcePartOne(prefix, &s)
	configTwo := configOne + testAccVPCPeeringConnectionAzureResourcePartTwo(prefix, &s)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {
				Source:            "hashicorp/azurerm",
				VersionConstraint: "=3.30.0",
			},
			"azuread": {
				Source:            "hashicorp/azuread",
				VersionConstraint: "=2.30.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: configOne,
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttrSet("aiven_azure_vpc_peering_connection.peering_connection", "id"),
					// We can't check peering_connection state, because it's updated async and gets ACTIVE any time later

					// Azure resources
					resource.TestCheckResourceAttrSet("azurerm_resource_group.resource_group", "id"),
					resource.TestCheckResourceAttrSet("azurerm_virtual_network.virtual_network", "id"),
					resource.TestCheckResourceAttrSet("azuread_application.application", "id"),
					resource.TestCheckResourceAttrSet("azuread_service_principal.app_principal", "id"),
					resource.TestCheckResourceAttrSet("azuread_application_password.app_password", "id"),
					resource.TestCheckResourceAttrSet("azurerm_role_assignment.app_role_assignment", "id"),
					resource.TestCheckResourceAttrSet("azuread_service_principal.aiven_principal", "id"),
					resource.TestCheckResourceAttrSet("azurerm_role_definition.role_definition", "id"),
					resource.TestCheckResourceAttrSet("azurerm_role_assignment.aiven_role_assignment", "id"),
				),
			},
			importStateByName("aiven_project_vpc.vpc"),
			importStateByName("aiven_azure_vpc_peering_connection.peering_connection"),
			importStateByName("azurerm_resource_group.resource_group"),
			importStateByName("azurerm_virtual_network.virtual_network"),
			importStateByName("azuread_application.application"),
			importStateByName("azuread_service_principal.app_principal"),
			importStateByName("azurerm_role_assignment.app_role_assignment"),
			importStateByName("azuread_service_principal.aiven_principal"),
			importStateByName("azurerm_role_definition.role_definition"),
			importStateByName("azurerm_role_assignment.aiven_role_assignment"),
			// This test runs dynamic provider config
			// azurerm_virtual_network_peering can not be imported cause terraform doesn't work well with dynamic vars
			// https://github.com/hashicorp/terraform/issues/27934
			{
				Config: configTwo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aiven_azure_vpc_peering_connection.peering_connection", "id"),
					resource.TestCheckResourceAttrSet("azurerm_virtual_network_peering.network_peering", "id"),
				),
			},
		},
	})
}

// testAccVPCPeeringConnectionAzureResourcePartOne
// Based on https://aiven.io/docs/platform/howto/vnet-peering-azure
func testAccVPCPeeringConnectionAzureResourcePartOne(prefix string, s *azureSecrets) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
  tenant_id       = %[3]q
  subscription_id = %[4]q
}

provider "azuread" {
  tenant_id = %[3]q
}

data "aiven_project" "project" {
  project = %[2]q
}

data "azurerm_subscription" "subscription" {
  subscription_id = %[4]q
}

resource "aiven_project_vpc" "vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "azure-germany-westcentral"
  network_cidr = "192.168.1.0/24"

  timeouts {
    create = "15m"
  }
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

# 1. Log in with an Azure admin account
# Skip, no login required for terraform

# 2. Create application object
resource "azuread_application" "application" {
  display_name = "%[1]s-application"
  # Accounts in any organizational directory (multi-tenant)
  #  https://learn.microsoft.com/en-us/azure/active-directory/develop/supported-accounts-validation
  sign_in_audience = "AzureADandPersonalMicrosoftAccount"

  api {
    requested_access_token_version = 2
  }
}

# 3. Create a service principal for your app object
resource "azuread_service_principal" "app_principal" {
  application_id = azuread_application.application.application_id
}

# 4. Set a password for your app object
resource "azuread_application_password" "app_password" {
  application_object_id = azuread_application.application.object_id
}

# 5. Find the id properties of your virtual network
# Skip, we have values in the state

# 6. Grant your service principal permissions to peer
resource "azurerm_role_assignment" "app_role_assignment" {
  role_definition_name = "Network Contributor"
  principal_id         = azuread_service_principal.app_principal.object_id
  scope                = azurerm_virtual_network.virtual_network.id
}

# 7. Create a service principal for the Aiven application object
resource "azuread_service_principal" "aiven_principal" {
  application_id = %[5]q
}

# 8. Create a custom role for the Aiven application object
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

# 9. Assign the custom role to the Aiven service principal
resource "azurerm_role_assignment" "aiven_role_assignment" {
  role_definition_id = azurerm_role_definition.role_definition.role_definition_resource_id
  principal_id       = azuread_service_principal.aiven_principal.object_id
  scope              = azurerm_virtual_network.virtual_network.id

  depends_on = [
    azuread_service_principal.aiven_principal,
    azurerm_role_assignment.app_role_assignment
  ]
}

# 10. Find your AD tenant id
# Skip, it's in the env

# 11. Create a peering connection from the Aiven Project VPC
# 12. Wait for the Aiven platform to set up the connection
resource "aiven_azure_vpc_peering_connection" "peering_connection" {
  vpc_id                = aiven_project_vpc.vpc.id
  peer_resource_group   = azurerm_resource_group.resource_group.name
  azure_subscription_id = data.azurerm_subscription.subscription.subscription_id
  vnet_name             = azurerm_virtual_network.virtual_network.name
  peer_azure_app_id     = azuread_application.application.application_id
  peer_azure_tenant_id  = %[3]q

  depends_on = [
    azurerm_role_assignment.aiven_role_assignment
  ]
}

`, prefix, s.Project, s.TenantID, s.SubscriptionID, s.AivenAppID)
}

// testAccVPCPeeringConnectionAzureResourcePartTwo returns dynamically configured provider
// https://github.com/hashicorp/terraform/issues/25244
func testAccVPCPeeringConnectionAzureResourcePartTwo(prefix string, s *azureSecrets) string {
	return fmt.Sprintf(`
# 13. Create peering from your VNet to the Project VPC's VNet
provider "azurerm" {
  features {}
  alias                = "app"
  client_id            = azuread_application.application.application_id
  client_secret        = azuread_application_password.app_password.value
  subscription_id      = data.azurerm_subscription.subscription.subscription_id
  tenant_id            = %[2]q
  auxiliary_tenant_ids = [azuread_service_principal.aiven_principal.application_tenant_id]
}

resource "azurerm_virtual_network_peering" "network_peering" {
  provider                     = azurerm.app
  name                         = "%[1]s-network-peering"
  remote_virtual_network_id    = aiven_azure_vpc_peering_connection.peering_connection.state_info["to-network-id"]
  resource_group_name          = azurerm_resource_group.resource_group.name
  virtual_network_name         = azurerm_virtual_network.virtual_network.name
  allow_virtual_network_access = true
}
`, prefix, s.TenantID)
}
