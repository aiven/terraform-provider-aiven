package billing_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationBillingGroup(t *testing.T) {
	// Payment method ID is required for this test
	paymentMethodID := os.Getenv("AIVEN_PAYMENT_METHOD_ID")
	if paymentMethodID == "" {
		t.Skip("Skipping test due to missing AIVEN_PAYMENT_METHOD_ID environment variable")
	}

	var (
		name             = "aiven_organization_billing_group.billing_group"
		organizationName = acc.OrganizationName()
	)

	// Create a template builder factory with common configurations
	baseConfig := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

// Add common address configurations
resource "aiven_organization_address" "billing_address" {
  organization_id = data.aiven_organization.org.id
  address_lines   = ["123 Main St", "Suite 456"]
  city            = "Helsinki"
  company_name    = "Test Company"
  country_code    = "FI"
  state           = "Uusimaa"
  zip_code        = "00100"
}
resource "aiven_organization_address" "shipping_address" {
  organization_id = data.aiven_organization.org.id
  address_lines   = ["456 Market St", "Floor 3"]
  city            = "San Francisco"
  company_name    = "Test Company"
  country_code    = "US"
  state           = "CA"
  zip_code        = "94105"
}
	`, organizationName)

	// Initial configuration
	initialConfig := baseConfig + fmt.Sprintf(`
resource "aiven_organization_billing_group" "billing_group" {
  organization_id        = data.aiven_organization.org.id
  billing_address_id     = aiven_organization_address.billing_address.address_id
  shipping_address_id    = aiven_organization_address.shipping_address.address_id
  billing_currency       = "USD"
  billing_group_name     = "Test Billing Group"
  custom_invoice_text    = "Custom invoice text"
  payment_method_id      = %q
  vat_id                 = "VAT123"
  billing_emails         = [{ email = "invoices@example.com" }]
  billing_contact_emails = [{ email = "billing@example.com" }]
}
	`, paymentMethodID)

	// Updated configuration
	updatedConfig := baseConfig + fmt.Sprintf(`
resource "aiven_organization_billing_group" "billing_group" {
  organization_id     = data.aiven_organization.org.id
  billing_address_id  = aiven_organization_address.billing_address.address_id
  shipping_address_id = aiven_organization_address.shipping_address.address_id
  billing_currency    = "EUR"
  billing_group_name  = "Updated Billing Group"
  custom_invoice_text = "Updated invoice text"
  payment_method_id   = %q
  vat_id              = "VAT456"
  billing_contact_emails = [
    { email = "billing@example.com" },
    { email = "billing2@example.com" },
  ]
  billing_emails = [
    { email = "invoices@example.com" },
    { email = "invoices2@example.com" },
  ]
}
	`, paymentMethodID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test creation with all fields
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttrSet(name, "billing_group_id"),

					// Check organization_id is correctly set
					resource.TestCheckResourceAttrPair(name, "organization_id", "data.aiven_organization.org", "id"),

					// Check address IDs are correctly set
					resource.TestCheckResourceAttrPair(name, "billing_address_id", "aiven_organization_address.billing_address", "address_id"),
					resource.TestCheckResourceAttrPair(name, "shipping_address_id", "aiven_organization_address.shipping_address", "address_id"),

					// Check email lists
					resource.TestCheckResourceAttr(name, "billing_contact_emails.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_contact_emails.*", map[string]string{"email": "billing@example.com"}),
					resource.TestCheckResourceAttr(name, "billing_emails.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_emails.*", map[string]string{"email": "invoices@example.com"}),

					// Check other fields
					resource.TestCheckResourceAttr(name, "payment_method_id", paymentMethodID),
					resource.TestCheckResourceAttr(name, "billing_currency", "USD"),
					resource.TestCheckResourceAttr(name, "billing_group_name", "Test Billing Group"),
					resource.TestCheckResourceAttr(name, "custom_invoice_text", "Custom invoice text"),
					resource.TestCheckResourceAttr(name, "vat_id", "VAT123"),
				),
			},
			{
				// Test update
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					// Check ID remains the same

					resource.TestCheckResourceAttrSet(name, "id"),

					// Check email lists
					resource.TestCheckResourceAttr(name, "billing_contact_emails.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_contact_emails.*", map[string]string{"email": "billing@example.com"}),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_contact_emails.*", map[string]string{"email": "billing2@example.com"}),
					resource.TestCheckResourceAttr(name, "billing_emails.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_emails.*", map[string]string{"email": "invoices@example.com"}),
					resource.TestCheckTypeSetElemNestedAttrs(name, "billing_emails.*", map[string]string{"email": "invoices2@example.com"}),

					// Check updated fields
					resource.TestCheckResourceAttr(name, "billing_currency", "EUR"),
					resource.TestCheckResourceAttr(name, "billing_group_name", "Updated Billing Group"),
					resource.TestCheckResourceAttr(name, "custom_invoice_text", "Updated invoice text"),
					resource.TestCheckResourceAttr(name, "vat_id", "VAT456"),
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
