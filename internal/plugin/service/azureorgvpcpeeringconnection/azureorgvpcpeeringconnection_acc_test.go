package azureorgvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/kelseyhightower/envconfig"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	azureOrgVPCPeeringResource = "aiven_azure_org_vpc_peering_connection"
	organizationVPCResource    = "aiven_organization_vpc"
)

func TestAccAivenAzureOrgVPCPeeringConnection(t *testing.T) {
	var (
		orgName      = acc.OrganizationName()
		templBuilder = template.InitializeTemplateStore(t).NewBuilder().
				AddDataSource("aiven_organization", map[string]any{
				"resource_name": "foo",
				"name":          orgName,
			})
		subscriptionID = "00000000-0000-0000-0000-000000000000"
		vnetName       = "test-vnet"
		resourceGroup  = "test-rg"
		appID          = "00000000-0000-0000-0000-000000000000"
		tenantID       = "00000000-0000-0000-0000-000000000000"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAzureOrgVPCPeeringResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "azure-germany-westcentral",
						"network_cidr":    "10.0.0.0/24",
					}).
					AddResource(azureOrgVPCPeeringResource, map[string]any{
						"resource_name":         "test_org_vpc_peering",
						"organization_id":       template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id":   template.Reference("aiven_organization_vpc.test_org_vpc.organization_vpc_id"),
						"azure_subscription_id": template.Literal(subscriptionID),
						"vnet_name":             template.Literal(vnetName),
						"peer_resource_group":   template.Literal(resourceGroup),
						"peer_azure_app_id":     template.Literal(appID),
						"peer_azure_tenant_id":  template.Literal(tenantID),
					}).MustRender(t),
				ExpectError: regexp.MustCompile(`REJECTED_BY_PEER`), // Azure app ID is invalid
			},
		},
	})
}

func TestAccAivenAzureOrgVPCPeeringConnection_backwardCompat(t *testing.T) {
	var config acc.AzurePeeringConfig
	if err := envconfig.Process("", &config); err != nil {
		t.Skipf("Azure test configuration is missing: %s", err)
	}
	prefix := acc.RandName("org-peering")
	resourceName := azureOrgVPCPeeringResource + ".test_peering"
	fixture := acc.AzurePeeringFixture(prefix, &config) + fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %q
}
resource "aiven_organization_vpc" "example" {
  organization_id = data.aiven_organization.foo.id
  cloud_name      = "azure-germany-westcentral"
  network_cidr    = "192.168.1.0/24"
}
`, acc.OrganizationName())
	peering := fixture + fmt.Sprintf(`
resource "aiven_azure_org_vpc_peering_connection" "test_peering" {
  organization_id       = data.aiven_organization.foo.id
  organization_vpc_id   = aiven_organization_vpc.example.organization_vpc_id
  azure_subscription_id = data.azurerm_subscription.subscription.subscription_id
  vnet_name             = azurerm_virtual_network.virtual_network.name
  peer_resource_group   = azurerm_resource_group.resource_group.name
  peer_azure_app_id     = azuread_application.application.application_id
  peer_azure_tenant_id  = %q
  depends_on            = [azurerm_role_assignment.aiven_role_assignment]
}
data "aiven_azure_org_vpc_peering_connection" "test_peering" {
  organization_id       = aiven_azure_org_vpc_peering_connection.test_peering.organization_id
  organization_vpc_id   = aiven_azure_org_vpc_peering_connection.test_peering.organization_vpc_id
  azure_subscription_id = aiven_azure_org_vpc_peering_connection.test_peering.azure_subscription_id
  vnet_name             = aiven_azure_org_vpc_peering_connection.test_peering.vnet_name
  peer_resource_group   = aiven_azure_org_vpc_peering_connection.test_peering.peer_resource_group
}
`, config.TenantID)
	var peeringState *terraform.State
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.foo", "id"),
		resource.TestCheckResourceAttrPair(resourceName, "organization_vpc_id", "aiven_organization_vpc.example", "organization_vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "azure_subscription_id", config.SubscriptionID),
		resource.TestCheckResourceAttrPair(resourceName, "vnet_name", "azurerm_virtual_network.virtual_network", "name"),
		resource.TestCheckResourceAttrPair(resourceName, "peer_resource_group", "azurerm_resource_group.resource_group", "name"),
		resource.TestCheckResourceAttrPair(resourceName, "peer_azure_app_id", "azuread_application.application", "application_id"),
		resource.TestCheckResourceAttr(resourceName, "peer_azure_tenant_id", config.TenantID),
		resource.TestCheckResourceAttrSet(resourceName, "peering_connection_id"),
		resource.TestCheckResourceAttr(resourceName, "state", "PENDING_PEER"),
		resource.TestCheckResourceAttrPair("data."+resourceName, "id", resourceName, "id"),
		func(state *terraform.State) error { peeringState = state; return nil },
	)
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig: peering, OldProviderVersion: "4.60.0", Checks: checks,
	})
	providers := map[string]resource.ExternalProvider{
		"azurerm": {Source: "hashicorp/azurerm", VersionConstraint: "=3.30.0"},
		"azuread": {Source: "hashicorp/azuread", VersionConstraint: "=2.30.0"},
	}
	for i := range steps {
		if steps[i].ExternalProviders == nil {
			steps[i].ExternalProviders = make(map[string]resource.ExternalProvider)
		}
		for name, provider := range providers {
			steps[i].ExternalProviders[name] = provider
		}
	}
	var identityImport acc.IdentityImportVersionCheck
	steps[1].ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}
	steps = append(steps, acc.ProviderRollbackSteps(t, steps[0], steps[1], &identityImport)...)
	steps = append(steps,
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, ExternalProviders: providers,
			ResourceName: resourceName, ImportState: true, ImportStateVerify: true,
		},
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, ExternalProviders: providers,
			ResourceName: resourceName, ImportState: true,
			ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
		},
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, ExternalProviders: providers,
			Config: fixture,
			// Check deletion while the parent VPC is still present.
			Check: func(_ *terraform.State) error { return testAccCheckAzureOrgVPCPeeringResourceDestroy(peeringState) },
		},
		// Reuse the VPC and Azure setup to exercise Create with the current provider too.
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, ExternalProviders: providers,
			Config: peering, Check: checks,
		},
	)
	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acc.TestAccPreCheck(t) },
		CheckDestroy:           testAccCheckAzureOrgVPCPeeringResourceDestroy,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		Steps:                  steps,
	})
}

func testAccCheckAzureOrgVPCPeeringResourceDestroy(s *terraform.State) error {
	ctx := context.Background()

	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error initializing Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != azureOrgVPCPeeringResource {
			continue
		}

		orgID, vpcID, cloudAccount, vnetName, resourceGroup, err := schemautil.SplitResourceID5(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting resource with ID: %q - %w", rs.Primary.ID, err)
		}

		orgVPC, err := c.OrganizationVpcGet(ctx, orgID, vpcID)
		if common.IsCritical(err) {
			return fmt.Errorf("error fetching VPC (%q): %w", vpcID, err)
		}

		if orgVPC == nil {
			return nil // Peering connection was deleted with the VPC
		}

		var pc *organizationvpc.OrganizationVpcGetPeeringConnectionOut
		for i := range orgVPC.PeeringConnections {
			pCon := &orgVPC.PeeringConnections[i]
			if pCon.PeerCloudAccount == cloudAccount &&
				pCon.PeerVpc == vnetName &&
				pCon.PeerResourceGroup == resourceGroup {
				pc = pCon
				break
			}
		}

		if pc != nil && pc.State != "DELETED" {
			connectionID := "<missing>"
			if pc.PeeringConnectionId != nil {
				connectionID = *pc.PeeringConnectionId
			}
			return fmt.Errorf("peering connection %q still exists", connectionID)
		}
	}

	return nil
}
