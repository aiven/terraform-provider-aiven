package address_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationAddress(t *testing.T) {
	deps := acc.CommonTestDependencies(t)

	var (
		name             = "aiven_organization_address.address"
		organizationName = deps.OrganizationName()
		templateStore    = template.InitializeTemplateStore(t)
	)

	// Create a template builder factory with the org data source already configured
	templBuilder := templateStore.NewBuilder().
		AddDataSource("aiven_organization", map[string]interface{}{
			"resource_name": "org",
			"name":          organizationName,
		}).Factory()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test creation with all fields
				Config: templBuilder().AddResource("aiven_organization_address", map[string]any{
					"resource_name":   "address",
					"organization_id": template.Reference("data.aiven_organization.org.id"),
					"address_lines":   []string{"123 Main St", "Suite 456"},
					"city":            "Helsinki",
					"company_name":    "Test Company",
					"country_code":    "FI",
					"state":           "Uusimaa",
					"zip_code":        "00100",
				}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttrSet(name, "address_id"),
					resource.TestCheckResourceAttrSet(name, "create_time"),
					resource.TestCheckResourceAttrSet(name, "update_time"),

					// Check organization_id is correctly set
					resource.TestCheckResourceAttrPair(name, "organization_id", "data.aiven_organization.org", "id"),

					// Check address_lines list
					resource.TestCheckResourceAttr(name, "address_lines.#", "2"),
					resource.TestCheckResourceAttr(name, "address_lines.0", "123 Main St"),
					resource.TestCheckResourceAttr(name, "address_lines.1", "Suite 456"),

					// Check other fields
					resource.TestCheckResourceAttr(name, "city", "Helsinki"),
					resource.TestCheckResourceAttr(name, "company_name", "Test Company"),
					resource.TestCheckResourceAttr(name, "country_code", "FI"),
					resource.TestCheckResourceAttr(name, "state", "Uusimaa"),
					resource.TestCheckResourceAttr(name, "zip_code", "00100"),
				),
			},
			{
				// Test update
				Config: templBuilder().AddResource("aiven_organization_address", map[string]any{
					"resource_name":   "address",
					"organization_id": template.Reference("data.aiven_organization.org.id"),
					"address_lines":   []string{"456 Market St", "Floor 3"},
					"city":            "San Francisco",
					"company_name":    "Updated Company",
					"country_code":    "US",
					"state":           "CA",
					"zip_code":        "94105",
				}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check ID remains the same
					resource.TestCheckResourceAttrSet(name, "id"),

					// Check updated fields
					resource.TestCheckResourceAttr(name, "address_lines.#", "2"),
					resource.TestCheckResourceAttr(name, "address_lines.0", "456 Market St"),
					resource.TestCheckResourceAttr(name, "address_lines.1", "Floor 3"),
					resource.TestCheckResourceAttr(name, "city", "San Francisco"),
					resource.TestCheckResourceAttr(name, "company_name", "Updated Company"),
					resource.TestCheckResourceAttr(name, "country_code", "US"),
					resource.TestCheckResourceAttr(name, "state", "CA"),
					resource.TestCheckResourceAttr(name, "zip_code", "94105"),
				),
			},
			{
				// Test import functionality
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
