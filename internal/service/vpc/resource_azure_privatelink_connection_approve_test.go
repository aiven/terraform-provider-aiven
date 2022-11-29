package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

type azureSecrets struct {
	Project        string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	AivenAppID     string `envconfig:"AIVEN_AZURE_APP_ID" required:"true"`
	TenantID       string `envconfig:"AZURE_TENANT_ID" required:"true"`
	SubscriptionID string `envconfig:"AZURE_SUBSCRIPTION_ID" required:"true"`
}

func TestAccAivenAzurePrivateLinkConnectionApproval_basic(t *testing.T) {
	var s azureSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	prefix := "test-tf-acc-plapproval-" + acctest.RandString(7)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {
				Source:            "hashicorp/azurerm",
				VersionConstraint: "=3.30.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAzurePrivateLinkConnectionApprovalResource(prefix, &s),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.project_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_pg.pg", "state", "RUNNING"),
					resource.TestCheckResourceAttr("aiven_azure_privatelink.private_link", "state", "active"),
					resource.TestCheckResourceAttr("aiven_azure_privatelink_connection_approval.approval", "state", "active"),
					resource.TestCheckResourceAttrSet("aiven_azure_privatelink_connection_approval.approval", "endpoint_ip_address"),
					resource.TestCheckResourceAttrSet("aiven_azure_privatelink_connection_approval.approval", "privatelink_connection_id"),

					// Azure resources
					resource.TestCheckResourceAttrSet("azurerm_resource_group.resource_group", "id"),
					resource.TestCheckResourceAttrSet("azurerm_virtual_network.virtual_network", "id"),
					resource.TestCheckResourceAttrSet("azurerm_subnet.subnet", "id"),
					resource.TestCheckResourceAttrSet("azurerm_private_endpoint.private_endpoint", "id"),
				),
			},
			importStateByName("aiven_project_vpc.project_vpc"),
			importStateByName("aiven_pg.pg"),
			importStateByName("aiven_azure_privatelink.private_link"),
			importStateByName("aiven_azure_privatelink_connection_approval.approval"),
			importStateByName("azurerm_resource_group.resource_group"),
			importStateByName("azurerm_virtual_network.virtual_network"),
			importStateByName("azurerm_subnet.subnet"),
			importStateByName("azurerm_private_endpoint.private_endpoint"),
		},
	})
}

func testAccAzurePrivateLinkConnectionApprovalResource(prefix string, s *azureSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "azurerm" {
  features {}
  tenant_id       = %[3]q
  subscription_id = %[4]q
}

resource "aiven_project_vpc" "project_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "azure-germany-north"
  network_cidr = "192.168.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aiven_static_ip" "static_ips" {
  count      = 2
  project    = aiven_project_vpc.project_vpc.project
  cloud_name = aiven_project_vpc.project_vpc.cloud_name
}

resource "aiven_pg" "pg" {
  service_name   = "%[1]s-pg"
  project        = data.aiven_project.project.project
  project_vpc_id = aiven_project_vpc.project_vpc.id
  cloud_name     = aiven_project_vpc.project_vpc.cloud_name
  plan           = "startup-4"
  static_ips     = [for sip in aiven_static_ip.static_ips : sip.static_ip_address_id]

  pg_user_config {
    static_ips = true

    privatelink_access {
      pg        = true
      pgbouncer = true
    }
  }
}

resource "aiven_azure_privatelink" "private_link" {
  project      = data.aiven_project.project.project
  service_name = aiven_pg.pg.service_name

  user_subscription_ids = [
    %[4]q,
  ]
}

resource "azurerm_resource_group" "resource_group" {
  location = "germanynorth"
  name     = "%[1]s-private-link"
}

resource "azurerm_virtual_network" "virtual_network" {
  name                = "%[1]s-virtual-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_subnet" "subnet" {
  name                 = "%[1]s-subnet"
  resource_group_name  = azurerm_resource_group.resource_group.name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_private_endpoint" "private_endpoint" {
  name                = "%[1]s-private-endpoint"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  subnet_id           = azurerm_subnet.subnet.id

  private_service_connection {
    name                           = aiven_pg.pg.service_name
    request_message                = aiven_pg.pg.service_name
    private_connection_resource_id = aiven_azure_privatelink.private_link.azure_service_id
    is_manual_connection           = true
  }

  depends_on = [
    aiven_azure_privatelink.private_link,
  ]
}

resource "aiven_azure_privatelink_connection_approval" "approval" {
  project             = data.aiven_project.project.project
  service_name        = aiven_pg.pg.service_name
  endpoint_ip_address = azurerm_private_endpoint.private_endpoint.private_service_connection[0].private_ip_address
}`, prefix, s.Project, s.TenantID, s.SubscriptionID)
}

func importStateByName(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName: name,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			root := s.RootModule()
			rs, ok := root.Resources[name]
			if !ok {
				return "", fmt.Errorf(`resource %q not found in the state`, name)
			}
			return rs.Primary.ID, nil
		},
	}
}
