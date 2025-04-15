package organizationaddress_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationAddress(t *testing.T) {
	var (
		name             = "aiven_organization_address.address"
		organizationName = acc.OrganizationName()
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

// TestAccAivenOrganizationAddress_CompanyNameHandling tests the workaround for company_name handling
// This test and the corresponding workaround in the code should be removed once the API is fixed
// to properly handle null vs. empty string for company_name.
func TestAccAivenOrganizationAddress_CompanyNameHandling(t *testing.T) {
	t.Skip("The company_name can't bet updated to null. Todo: run this test when the API is fixed.")

	var (
		name             = "aiven_organization_address.test_address"
		organizationName = acc.OrganizationName()
		templateStore    = template.InitializeTemplateStore(t)
	)

	// Create base configuration for all test steps
	baseConfig := map[string]any{
		"resource_name":   "test_address",
		"organization_id": template.Reference("data.aiven_organization.org.id"),
		"address_lines":   []string{"123 Test St"},
		"city":            "Test City",
		"country_code":    "US",
		"state":           "CA",
		"zip_code":        "12345",
		// company_name is intentionally omitted
	}

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
				// Step 1: Create address without company_name
				Config: templBuilder().AddResource("aiven_organization_address", baseConfig).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					// Verify company_name is not set (null)
					resource.TestCheckNoResourceAttr(name, "company_name"),
				),
			},
			{
				// Step 2: Apply the same configuration to check for drift
				// This verifies our workaround for handling empty strings from API
				Config:   templBuilder().AddResource("aiven_organization_address", baseConfig).MustRender(t),
				PlanOnly: true, // Only check the plan, no changes should be planned
			},
			{
				// Step 3: Update to add a company_name
				Config: templBuilder().AddResource("aiven_organization_address",
					mergeConfigs(baseConfig, map[string]any{
						"company_name": "Test Company", // Now adding company_name
					})).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Verify company_name is now set
					resource.TestCheckResourceAttr(name, "company_name", "Test Company"),
				),
			},
			{
				PreConfig: func() {
					baseConfig["state"] = "New State" // Change state to trigger update
				},
				// Step 4: Update to remove company_name again
				Config: templBuilder().AddResource("aiven_organization_address", baseConfig).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Verify company_name is not set again (null)
					resource.TestCheckNoResourceAttr(name, "company_name"),
				),
			},
		},
	})
}

// mergeConfigs merges a base configuration with an override configuration
func mergeConfigs(base, override map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy base values
	for k, v := range base {
		result[k] = v
	}

	// Override with new values
	for k, v := range override {
		result[k] = v
	}

	return result
}

func TestAccAivenOrganizationAddressDataSource(t *testing.T) {
	var (
		organizationName = acc.OrganizationName()
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
