package billing_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationBillingGroupListDataSource(t *testing.T) {
	deps := acc.CommonTestDependencies(t)

	var (
		organizationName = deps.OrganizationName()
		dataSourceName   = "data.aiven_organization_billing_group_list.ds"
		resourceName     = "aiven_organization_billing_group.group"
		templateStore    = template.InitializeTemplateStore(t)
	)

	// Create a template builder factory with the org data source already configured
	templBuilder := templateStore.NewBuilder().
		AddDataSource("aiven_organization", map[string]interface{}{
			"resource_name": "org",
			"name":          organizationName,
		}).Factory()

	config := templBuilder().
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
		AddResource("aiven_organization_billing_group", map[string]any{
			"resource_name":          "group",
			"organization_id":        template.Reference("data.aiven_organization.org.id"),
			"billing_group_name":     "Test Billing Group",
			"billing_currency":       "EUR",
			"billing_emails":         []string{"test@example.com"},
			"billing_contact_emails": []string{"contact@example.com"},
			"billing_address_id":     "test-address-id",
			"shipping_address_id":    "test-shipping-id",
			"vat_id":                 "TEST123456",
			"custom_invoice_text":    "Test Invoice Text",
		}).
		AddDataSource("aiven_organization_billing_group_list", map[string]any{
			"resource_name":   "ds",
			"organization_id": template.Reference("data.aiven_organization.org.id"),
		}).MustRender(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a billing group and verify it appears in the list
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_groups.#", "1"),

					// Check the first billing group matches our resource
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.organization_id", resourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_group_id", resourceName, "billing_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_group_name", resourceName, "billing_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_currency", resourceName, "billing_currency"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_address_id", resourceName, "billing_address_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.shipping_address_id", resourceName, "shipping_address_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.vat_id", resourceName, "vat_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.custom_invoice_text", resourceName, "custom_invoice_text"),

					// Check email lists
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_emails.#", resourceName, "billing_emails.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_emails.0", resourceName, "billing_emails.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_contact_emails.#", resourceName, "billing_contact_emails.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_groups.0.billing_contact_emails.0", resourceName, "billing_contact_emails.0"),

					// Directly check some values for additional confirmation
					resource.TestCheckResourceAttr(dataSourceName, "billing_groups.0.billing_group_name", "Test Billing Group"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_groups.0.billing_currency", "EUR"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_groups.0.billing_emails.0", "test@example.com"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_groups.0.billing_contact_emails.0", "contact@example.com"),
				),
			},
		},
	})
}
