package organizationvpc_test

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

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	organizationVPCResource = "aiven_organization_vpc"
)

func TestAccAivenOrganizationVPC(t *testing.T) {
	var identityImport acc.IdentityImportVersionCheck
	var (
		orgName = acc.OrganizationName()

		templBuilder = template.InitializeTemplateStore(t).NewBuilder().
				AddDataSource("aiven_organization", map[string]any{
				"resource_name": "foo",
				"name":          orgName,
			}).Factory()

		resourceName   = fmt.Sprintf("%s.%s", organizationVPCResource, "test_org_vpc")
		dataSourceName = fmt.Sprintf("data.%s.%s", organizationVPCResource, "vpc_ds")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationVPCResourceDestroy,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		Steps: []resource.TestStep{
			{
				// invalid CIDR range
				Config: `
resource "aiven_organization_vpc" "test_validation" {
  organization_id = "test-org-id"
  cloud_name      = "aws-eu-west-1"
  network_cidr    = "256.256.256.256/24"
}`,
				ExpectError: regexp.MustCompile(`must be a valid CIDR Value`),
			},
			{
				// basic VPC creation
				Config: templBuilder().
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "aws-eu-west-1",
						"network_cidr":    "10.0.0.0/24",
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "aws-eu-west-1"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", string(organizationvpc.OrganizationVpcStateTypeActive)),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
				),
			},
			{
				// test display_name update
				Config: templBuilder().
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "aws-eu-west-1",
						"network_cidr":    "10.0.0.0/24",
						"display_name":    "updated organization VPC",
					}).MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "aws-eu-west-1"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "updated organization VPC"),
					resource.TestCheckResourceAttr(resourceName, "state", string(organizationvpc.OrganizationVpcStateTypeActive)),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
				),
			},
			{
				// test ForceNew on network_cidr change
				Config: templBuilder().
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "aws-eu-west-1",
						"network_cidr":    "10.1.0.0/24", // Changed CIDR
					}).MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "aws-eu-west-1"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "10.1.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", string(organizationvpc.OrganizationVpcStateTypeActive)),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
				),
			},
			{
				// import state
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				SkipFunc:        identityImport.Skip,
			},
			{
				// test the data source
				Config: templBuilder().
					AddResource(organizationVPCResource, map[string]any{
						"resource_name":   "test_org_vpc",
						"organization_id": template.Reference("data.aiven_organization.foo.id"),
						"cloud_name":      "aws-eu-west-1",
						"network_cidr":    "10.1.0.0/24",
					}).
					AddDataSource(organizationVPCResource, map[string]any{
						"resource_name":       "vpc_ds",
						"organization_id":     template.Reference("data.aiven_organization.foo.id"),
						"organization_vpc_id": template.Reference(fmt.Sprintf("%s.organization_vpc_id", resourceName)),
					}).MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_id", resourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cloud_name", resourceName, "cloud_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_cidr", resourceName, "network_cidr"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttr(dataSourceName, "state", string(organizationvpc.OrganizationVpcStateTypeActive)),
					resource.TestCheckResourceAttrSet(dataSourceName, "organization_vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_vpc_id", resourceName, "organization_vpc_id"),
				),
			},
		},
	})
}

func TestAccAivenOrganizationVPC_backwardCompat(t *testing.T) {
	resourceName := "aiven_organization_vpc.test_org_vpc"
	config := testAccOrganizationVPCResourceOnly(acc.OrganizationName(), "aws-eu-west-1", "10.2.0.0/24")
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "cloud_name", "aws-eu-west-1"),
		resource.TestCheckResourceAttr(resourceName, "network_cidr", "10.2.0.0/24"),
		resource.TestCheckResourceAttr(resourceName, "state", string(organizationvpc.OrganizationVpcStateTypeActive)),
		resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
		resource.TestCheckResourceAttrSet(resourceName, "organization_vpc_id"),
	)

	testUpgrade := func(t *testing.T, oldVersion string) {
		t.Helper()
		var identityImport acc.IdentityImportVersionCheck
		testCase := resource.TestCase{
			PreCheck:               func() { acc.TestAccPreCheck(t) },
			CheckDestroy:           testAccCheckOrganizationVPCResourceDestroy,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
			Steps: []resource.TestStep{
				{
					ExternalProviders: map[string]resource.ExternalProvider{
						"aiven": acc.ExternalAivenProvider(t, oldVersion),
					},
					Config: config,
					Check:  checks,
				},
				{
					ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
					Config:                   config,
					Check:                    checks,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					}},
				},
				{
					ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
					ResourceName:             resourceName,
					ImportState:              true,
					ImportStateKind:          resource.ImportBlockWithResourceIdentity,
					SkipFunc:                 identityImport.Skip,
				},
			},
		}
		identityImportStep := testCase.Steps[2]
		testCase.Steps = append(testCase.Steps[:2], acc.ProviderRollbackSteps(t, testCase.Steps[0], testCase.Steps[1], &identityImport)...)
		testCase.Steps = append(testCase.Steps, identityImportStep)
		resource.Test(t, testCase)
	}
	t.Run("SDK", func(t *testing.T) { testUpgrade(t, "4.56.0") })
	t.Run("Framework", func(t *testing.T) { testUpgrade(t, "4.65.1") })
}

func testAccOrganizationVPCResourceOnly(orgName, cloud, cidr string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

resource "aiven_organization_vpc" "test_org_vpc" {
  organization_id = data.aiven_organization.foo.id
  cloud_name      = %[2]q
  network_cidr    = %[3]q
}
`, orgName, cloud, cidr)
}

func testAccCheckOrganizationVPCResourceDestroy(s *terraform.State) error {
	ctx := context.Background()

	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error initializing Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != organizationVPCResource {
			continue
		}

		orgID, vpcID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting resource ID: %w", err)
		}

		orgVPC, err := c.OrganizationVpcGet(ctx, orgID, vpcID)
		if common.IsCritical(err) {
			return fmt.Errorf("error fetching VPC (%q): %w", vpcID, err)
		}

		if orgVPC != nil {
			return fmt.Errorf("VPC (%q) for organization (%q) still exists", vpcID, orgID)
		}
	}

	return nil
}
