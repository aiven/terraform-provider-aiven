package billing_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationBillingGroupDataSource(t *testing.T) {
	// Payment method ID is required for this test
	paymentMethodID := os.Getenv("AIVEN_PAYMENT_METHOD_ID")
	if paymentMethodID == "" {
		t.Skip("Skipping test due to missing AIVEN_PAYMENT_METHOD_ID environment variable")
	}

	var (
		organizationName = acc.OrganizationName()
		dataSourceName   = "data.aiven_organization_billing_group.ds"
		resourceName     = "aiven_organization_billing_group.billing_group"
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
			"address_lines":   []string{"456 Market St", "Floor 3"},
			"city":            "San Francisco",
			"company_name":    "Test Company",
			"country_code":    "US",
			"state":           "CA",
			"zip_code":        "94105",
		}).
		AddResource("aiven_organization_billing_group", map[string]any{
			"resource_name":          "billing_group",
			"organization_id":        template.Reference("data.aiven_organization.org.id"),
			"billing_address_id":     template.Reference("aiven_organization_address.address.address_id"),
			"billing_contact_emails": []string{"billing@example.com"},
			"billing_currency":       "USD",
			"billing_emails":         []string{"invoices@example.com"},
			"billing_group_name":     "Test Billing Group",
			"custom_invoice_text":    "Custom invoice text",
			"payment_method_id":      paymentMethodID,
			"shipping_address_id":    template.Reference("aiven_organization_address.address.address_id"),
			"vat_id":                 "VAT123",
		}).
		AddDataSource("aiven_organization_billing_group", map[string]any{
			"resource_name":    "ds",
			"organization_id":  template.Reference("data.aiven_organization.org.id"),
			"billing_group_id": template.Reference("aiven_organization_billing_group.billing_group.billing_group_id"),
		}).MustRender(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a resource and read it with data source
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),

					// Check all fields match resource
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_id", resourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_group_id", resourceName, "billing_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_address_id", resourceName, "billing_address_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payment_method_id", resourceName, "payment_method_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shipping_address_id", resourceName, "shipping_address_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_group_name", resourceName, "billing_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_currency", resourceName, "billing_currency"),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_invoice_text", resourceName, "custom_invoice_text"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vat_id", resourceName, "vat_id"),

					// Check email lists
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_contact_emails.#", resourceName, "billing_contact_emails.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_contact_emails.0", resourceName, "billing_contact_emails.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_emails.#", resourceName, "billing_emails.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_emails.0", resourceName, "billing_emails.0"),

					// Directly check some values for additional confirmation
					resource.TestCheckResourceAttr(dataSourceName, "billing_group_name", "Test Billing Group"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_currency", "USD"),
				),
			},
		},
	})
}
