package azureorgvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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
	acc.SkipIfNotBeta(t)
	t.Skip("Skipping due to Azure SDK dependency")

	env := acc.RequireEnvVars(
		t,
		"AZURE_SUBSCRIPTION_ID",
		"AZURE_VNET_NAME",
		"AZURE_RESOURCE_GROUP",
		"AZURE_APP_ID",
		"AZURE_TENANT_ID",
	)
	resourceName := azureOrgVPCPeeringResource + ".test_peering"
	config := testAccAzureOrgVPCPeeringBackwardCompatConfig(acc.OrganizationName(), env)
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
		resource.TestCheckResourceAttrSet(resourceName, "organization_vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "azure_subscription_id", env["AZURE_SUBSCRIPTION_ID"]),
		resource.TestCheckResourceAttr(resourceName, "vnet_name", env["AZURE_VNET_NAME"]),
		resource.TestCheckResourceAttr(resourceName, "peer_resource_group", env["AZURE_RESOURCE_GROUP"]),
		resource.TestCheckResourceAttr(resourceName, "peer_azure_app_id", env["AZURE_APP_ID"]),
		resource.TestCheckResourceAttr(resourceName, "peer_azure_tenant_id", env["AZURE_TENANT_ID"]),
		resource.TestCheckResourceAttrSet(resourceName, "peering_connection_id"),
		resource.TestCheckResourceAttrSet(resourceName, "state"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckAzureOrgVPCPeeringResourceDestroy,
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.60.0",
			Checks:             checks,
		}),
	})
}

func testAccAzureOrgVPCPeeringBackwardCompatConfig(orgName string, env map[string]string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

resource "aiven_organization_vpc" "example" {
  organization_id = data.aiven_organization.foo.id
  cloud_name      = "azure-germany-westcentral"
  network_cidr    = "10.0.0.0/24"
}

resource "aiven_azure_org_vpc_peering_connection" "test_peering" {
  organization_id       = data.aiven_organization.foo.id
  organization_vpc_id   = aiven_organization_vpc.example.organization_vpc_id
  azure_subscription_id = %[2]q
  vnet_name             = %[3]q
  peer_resource_group   = %[4]q
  peer_azure_app_id     = %[5]q
  peer_azure_tenant_id  = %[6]q
}
`,
		orgName,
		env["AZURE_SUBSCRIPTION_ID"],
		env["AZURE_VNET_NAME"],
		env["AZURE_RESOURCE_GROUP"],
		env["AZURE_APP_ID"],
		env["AZURE_TENANT_ID"],
	)
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

		if pc != nil {
			connectionID := "<missing>"
			if pc.PeeringConnectionId != nil {
				connectionID = *pc.PeeringConnectionId
			}
			return fmt.Errorf("peering connection %q still exists", connectionID)
		}
	}

	return nil
}
