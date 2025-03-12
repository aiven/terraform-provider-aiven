package address_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationAddressDataSource(t *testing.T) {
	deps := acc.CommonTestDependencies(t)

	var (
		organizationName = deps.OrganizationName()
		dataSourceName   = "data.aiven_organization_address.ds"
		resourceName     = "aiven_organization_address.address"
		templBuilder     = template.InitializeTemplateStore(t).NewBuilder().
					AddDataSource("aiven_organization", map[string]interface{}{
				"resource_name": "org",
				"name":          organizationName,
			}).Factory()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a resource and read it with data source
				Config: templBuilder().
					AddResource("aiven_organization_address", map[string]any{
						"resource_name":   "address",
						"organization_id": template.Reference("data.aiven_organization.org.id"),
						"address_lines":   []string{"123 Main St", "Suite 456"},
						"city":            "Helsinki",
						"company_name":    "Test Company",
						"country_code":    "FI",
						"state":           "Uusimaa",
						"zip_code":        "00100",
					}).
					AddDataSource("aiven_organization_address", map[string]any{
						"resource_name":   "ds",
						"organization_id": template.Reference("data.aiven_organization.org.id"),
						"address_id":      template.Reference("aiven_organization_address.address.address_id"),
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "create_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "update_time"),

					// Check all fields match resource
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_id", resourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_id", resourceName, "address_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "city", resourceName, "city"),
					resource.TestCheckResourceAttrPair(dataSourceName, "company_name", resourceName, "company_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "country_code", resourceName, "country_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zip_code", resourceName, "zip_code"),

					// Check address_lines
					resource.TestCheckResourceAttrPair(dataSourceName, "address_lines.#", resourceName, "address_lines.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_lines.0", resourceName, "address_lines.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_lines.1", resourceName, "address_lines.1"),

					// Directly check some values for additional confirmation
					resource.TestCheckResourceAttr(dataSourceName, "city", "Helsinki"),
					resource.TestCheckResourceAttr(dataSourceName, "country_code", "FI"),
				),
			},
		},
	})
}
