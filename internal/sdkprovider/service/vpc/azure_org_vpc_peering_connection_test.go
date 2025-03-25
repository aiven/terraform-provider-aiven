package vpc_test

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
)

func TestAccAivenAzureOrgVPCPeeringConnection(t *testing.T) {
	t.Skip("Skipping due to Azure SDK dependency")

	var (
		orgName      = acc.OrganizationName()
		templBuilder = template.InitializeTemplateStore(t).NewBuilder().
				AddDataSource("aiven_organization", map[string]interface{}{
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
				ExpectError: regexp.MustCompile(`peer_azure_app_id '.*' does not refer to a valid application object`), // Azure app ID is invalid
			},
		},
	})
}

func testAccCheckAzureOrgVPCPeeringResourceDestroy(s *terraform.State) error {
	ctx := context.Background()

	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error initializing Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != awsOrgVPCPeeringResource {
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
		for _, pCon := range orgVPC.PeeringConnections {
			if pCon.PeerCloudAccount == cloudAccount && pCon.PeerVpc == vnetName && pCon.PeerResourceGroup == resourceGroup {
				pc = &pCon
				break
			}
		}

		if pc != nil {
			return fmt.Errorf("peering connection %q still exists", *pc.PeeringConnectionId)
		}
	}

	return nil
}
