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
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	azureOrgVPCPeeringResource = "aiven_azure_org_vpc_peering_connection"
)

func TestAccAivenAzureOrgVPCPeeringConnection(t *testing.T) {
	var (
		orgName        = acc.SkipIfEnvVarsNotSet(t, "AIVEN_ORGANIZATION_NAME")["AIVEN_ORGANIZATION_NAME"]
		registry       = preSetAzureOrgVPCPeeringTemplates(t)
		newComposition = func() *acc.CompositionBuilder {
			return registry.NewCompositionBuilder().
				Add("organization_data", map[string]any{
					"organization_name": orgName})
		}

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
				Config: newComposition().
					Add(organizationVPCResource, map[string]any{
						"resource_name": "test_org_vpc",
						"cloud_name":    "azure-germany-westcentral",
						"network_cidr":  "10.0.0.0/24",
					}).
					Add(azureOrgVPCPeeringResource, map[string]any{
						"resource_name":         "test_org_vpc_peering",
						"organization_id":       acc.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id":   acc.Reference("aiven_organization_vpc.test_org_vpc.organization_vpc_id"),
						"azure_subscription_id": acc.Literal(subscriptionID),
						"vnet_name":             acc.Literal(vnetName),
						"peer_resource_group":   acc.Literal(resourceGroup),
						"peer_azure_app_id":     acc.Literal(appID),
						"peer_azure_tenant_id":  acc.Literal(tenantID),
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

func preSetAzureOrgVPCPeeringTemplates(t *testing.T) *acc.TemplateRegistry {
	t.Helper()

	registry := acc.NewTemplateRegistry(azureOrgVPCPeeringResource)

	registry.MustAddTemplate(t, "organization_data", `
data "aiven_organization" "foo" {
  name = "{{ .organization_name }}"
}`)

	registry.MustAddTemplate(t, organizationVPCResource, `
resource "aiven_organization_vpc" "{{ .resource_name }}" {
    organization_id = data.aiven_organization.foo.id
    cloud_name     = "{{ .cloud_name }}"
    network_cidr   = "{{ .network_cidr }}"
}`)

	registry.MustAddTemplate(t, azureOrgVPCPeeringResource, `
resource "aiven_azure_org_vpc_peering_connection" "{{ .resource_name }}" {
	organization_id      	= {{ if .organization_id.IsLiteral }}"{{ .organization_id.Value }}"{{ else }}{{ .organization_id.Value }}{{ end }}
	organization_vpc_id   	= {{ if .organization_vpc_id.IsLiteral }}"{{ .organization_vpc_id.Value }}"{{ else }}{{ .organization_vpc_id.Value }}{{ end }}
	azure_subscription_id 	= {{ if .azure_subscription_id.IsLiteral }}"{{ .azure_subscription_id.Value }}"{{ else }}{{ .azure_subscription_id.Value }}{{ end }}
	vnet_name      			= {{ if .vnet_name.IsLiteral }}"{{ .vnet_name.Value }}"{{ else }}{{ .vnet_name.Value }}{{ end }}
	peer_resource_group 	= {{ if .peer_resource_group.IsLiteral }}"{{ .peer_resource_group.Value }}"{{ else }}{{ .peer_resource_group.Value }}{{ end }}
	peer_azure_app_id   	= {{ if .peer_azure_app_id.IsLiteral }}"{{ .peer_azure_app_id.Value }}"{{ else }}{{ .peer_azure_app_id.Value }}{{ end }}
	peer_azure_tenant_id 	= {{ if .peer_azure_tenant_id.IsLiteral }}"{{ .peer_azure_tenant_id.Value }}"{{ else }}{{ .peer_azure_tenant_id.Value }}{{ end }}
}`)

	return registry
}
