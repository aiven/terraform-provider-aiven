package billinggroup_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationBillingGroup(t *testing.T) {
	// Payment method ID is required for this test
	paymentMethodID := acc.PaymentMethodID()
	if paymentMethodID == "" {
		t.Skip("Skipping test due to missing AIVEN_PAYMENT_METHOD_ID environment variable")
	}

	var (
		name             = "aiven_organization_billing_group.billing_group"
		organizationName = acc.OrganizationName()
	)

	// Base configuration with shared resources:
	// - Organization data source
	// - Billing address in Helsinki, Finland
	// - Shipping address in San Francisco, USA
	baseConfig := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_organization_address" "billing_address" {
  organization_id = data.aiven_organization.org.id
  address_lines   = ["123 Main St", "Suite 456"]
  city            = "Helsinki"
  name            = "Test Company"
  country_code    = "FI"
  state           = "Uusimaa"
  zip_code        = "00100"
}

resource "aiven_organization_address" "shipping_address" {
  organization_id = data.aiven_organization.org.id
  address_lines   = ["456 Market St", "Floor 3"]
  city            = "San Francisco"
  name            = "Test Company"
  country_code    = "US"
  state           = "CA"
  zip_code        = "94105"
}`, organizationName)

	// Initial configuration creates a billing group with:
	// - USD currency
	// - Single billing contact email
	// - Single billing email
	// - Basic billing group name and invoice text
	initialConfig := baseConfig + fmt.Sprintf(`
resource "aiven_organization_billing_group" "billing_group" {
  organization_id        = data.aiven_organization.org.id
  billing_address_id     = aiven_organization_address.billing_address.address_id
  billing_contact_emails = ["billing@example.com"]
  currency               = "USD"
  billing_emails         = ["invoices@example.com"]
  billing_group_name     = "Test Billing Group"
  custom_invoice_text    = "Custom invoice text"
  payment_method_id      = %q
  shipping_address_id    = aiven_organization_address.shipping_address.address_id
  vat_id                 = "VAT123"
}`, paymentMethodID)

	// Updated configuration modifies the billing group with:
	// - Changed currency to EUR
	// - Added additional billing contact email
	// - Added additional billing email
	// - Updated billing group name and invoice text
	// - Updated VAT ID
	updatedConfig := baseConfig + fmt.Sprintf(`
resource "aiven_organization_billing_group" "billing_group" {
  organization_id        = data.aiven_organization.org.id
  billing_address_id     = aiven_organization_address.billing_address.address_id
  billing_contact_emails = ["billing@example.com", "billing2@example.com"]
  currency               = "EUR"
  billing_emails         = ["invoices@example.com", "invoices2@example.com"]
  billing_group_name     = "Updated Billing Group"
  custom_invoice_text    = "Updated invoice text"
  payment_method_id      = %q
  shipping_address_id    = aiven_organization_address.shipping_address.address_id
  vat_id                 = "VAT456"
}`, paymentMethodID)

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
					resource.TestCheckTypeSetElemAttr(name, "billing_contact_emails.*", "billing@example.com"),
					resource.TestCheckResourceAttr(name, "billing_emails.#", "1"),
					resource.TestCheckTypeSetElemAttr(name, "billing_emails.*", "invoices@example.com"),

					// Check other fields
					resource.TestCheckResourceAttr(name, "payment_method_id", paymentMethodID),
					resource.TestCheckResourceAttr(name, "currency", "USD"),
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
					resource.TestCheckTypeSetElemAttr(name, "billing_contact_emails.*", "billing2@example.com"),
					resource.TestCheckTypeSetElemAttr(name, "billing_contact_emails.*", "billing@example.com"),
					resource.TestCheckResourceAttr(name, "billing_emails.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, "billing_emails.*", "invoices2@example.com"),
					resource.TestCheckTypeSetElemAttr(name, "billing_emails.*", "invoices@example.com"),

					// Check updated fields
					resource.TestCheckResourceAttr(name, "currency", "EUR"),
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
