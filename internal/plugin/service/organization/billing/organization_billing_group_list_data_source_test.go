package billing_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationBillingGroupListDataSource(t *testing.T) {
	// Payment method ID is required for this test
	paymentMethodID := os.Getenv("AIVEN_PAYMENT_METHOD_ID")
	if paymentMethodID == "" {
		t.Skip("Skipping test due to missing AIVEN_PAYMENT_METHOD_ID environment variable")
	}

	var (
		organizationName = acc.OrganizationName()
		dataSourceName   = "data.aiven_organization_billing_group_list.ds"
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
			"billing_address_id":     template.Reference("aiven_organization_address.address.address_id"),
			"payment_method_id":      paymentMethodID,
			"shipping_address_id":    template.Reference("aiven_organization_address.address.address_id"),
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

					// Check that billing_groups list exists and has at least one item
					resource.TestCheckResourceAttrSet(dataSourceName, "billing_groups.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "billing_groups.0.billing_group_id"),
				),
			},
		},
	})
}
