package billinggrouplist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationBillingGroupListDataSource(t *testing.T) {
	// Payment method ID is required for this test
	paymentMethodID := acc.PaymentMethodID()
	if paymentMethodID == "" {
		t.Skip("Skipping test due to missing AIVEN_PAYMENT_METHOD_ID environment variable")
	}

	var (
		organizationName = acc.OrganizationName()
		dataSourceName   = "data.aiven_organization_billing_group_list.ds"
	)

	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_organization_address" "address" {
  organization_id = data.aiven_organization.org.id
  address_lines   = ["123 Main St", "Suite 456"]
  city            = "Helsinki"
  name            = "Test Company"
  country_code    = "FI"
  state           = "Uusimaa"
  zip_code        = "00100"
}

resource "aiven_organization_billing_group" "group" {
  organization_id        = data.aiven_organization.org.id
  billing_group_name     = "Test Billing Group"
  currency               = "EUR"
  billing_emails         = ["test@example.com"]
  billing_contact_emails = ["contact@example.com"]
  billing_address_id     = aiven_organization_address.address.address_id
  payment_method_id      = %q
  shipping_address_id    = aiven_organization_address.address.address_id
  vat_id                 = "TEST123456"
  custom_invoice_text    = "Test Invoice Text"
}

data "aiven_organization_billing_group_list" "ds" {
  organization_id = data.aiven_organization.org.id
}`, organizationName, paymentMethodID)

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
