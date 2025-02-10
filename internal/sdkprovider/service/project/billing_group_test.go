package project_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
)

func TestAccAivenBillingGroup_basic(t *testing.T) {
	resourceName := "aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenBillingGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "billing_extra_text", "test reference number 123"),
					resource.TestCheckResourceAttr(resourceName, "company", "Test Company Inc"),
					resource.TestCheckResourceAttr(resourceName, "address_lines.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "city", "Test City"),
					resource.TestCheckResourceAttr(resourceName, "zip_code", "12345"),
					resource.TestCheckResourceAttr(resourceName, "state", "Test State"),
					resource.TestCheckResourceAttr(resourceName, "vat_id", "TEST-VAT-123"),
					resource.TestCheckResourceAttr(resourceName, "billing_emails.#", "2"),
				),
			},
			{
				// Test importing billing group
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test updating billing group
				Config: testAccBillingGroupResourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_currency", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "billing_extra_text", "updated reference number 456"),
					resource.TestCheckResourceAttr(resourceName, "company", "Updated Company Ltd"),
					resource.TestCheckResourceAttr(resourceName, "address_lines.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "address_lines.*", "456 Updated Street"),
					resource.TestCheckTypeSetElemAttr(resourceName, "address_lines.*", "Suite 8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "address_lines.*", "Main Avenue"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "FI"),
					resource.TestCheckResourceAttr(resourceName, "city", "Updated City"),
					resource.TestCheckResourceAttr(resourceName, "zip_code", "54321"),
					resource.TestCheckResourceAttr(resourceName, "state", "Updated State"),
					resource.TestCheckResourceAttr(resourceName, "vat_id", "UPDATED-VAT-456"),
					resource.TestCheckResourceAttr(resourceName, "billing_emails.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "billing_emails.*", "ivan.savciuc+test2@aiven.fi"),
					resource.TestCheckTypeSetElemAttr(resourceName, "billing_emails.*", "ivan.savciuc+test3@aiven.fi"),
				),
			},
			{
				// Test removing optional set fields
				Config: testAccBillingGroupResourceOptionalFieldsRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "address_lines.#", "0"),
				),
			},
			{
				Config: testCopyBillingGroupFromExistingOne(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "name", fmt.Sprintf("copy-test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr("aiven_billing_group.bar", "billing_currency", "EUR"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "billing_currency", "EUR"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "city", "Helsinki"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "company", "Aiven Oy"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "vat_id", "abc"),
				),
			},
		},
	})
}

func testAccCheckAivenBillingGroupResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error getting aiven client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each billing group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_billing_group" {
			continue
		}

		bg, err := c.BillingGroupGet(ctx, rs.Primary.ID)
		var e avngen.Error
		if common.IsCritical(err) && errors.As(err, &e) && e.Status != 500 {
			return fmt.Errorf("error getting a billing group by id: %w", err)
		}

		if bg != nil {
			return fmt.Errorf("billing group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBillingGroupResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
    name                = "test-acc-bg-%[1]s"
    billing_emails      = ["ivan.savciuc+test1@aiven.fi", "ivan.savciuc+test2@aiven.fi"]
    billing_currency    = "USD"
    billing_extra_text  = "test reference number 123"
    company            = "Test Company Inc"
    address_lines      = ["123 Test Street", "Floor 4"]
    country_code       = "US"
    city              = "Test City"
    zip_code          = "12345"
    state             = "Test State"
    vat_id            = "TEST-VAT-123"
}

data "aiven_billing_group" "bar" {
  billing_group_id = aiven_billing_group.foo.id
}

resource "aiven_project" "pr1" {
  project       = "test-acc-pr-%[1]s"
  billing_group = aiven_billing_group.foo.id
  parent_id     = aiven_organization.foo.id

  depends_on = [aiven_billing_group.foo]
}`, name)
}

func testAccBillingGroupResourceOptionalFieldsRemoved(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
    name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
    name = "test-acc-bg-%[1]s"
}

data "aiven_billing_group" "bar" {
    billing_group_id = aiven_billing_group.foo.id
}

resource "aiven_project" "pr1" {
    project       = "test-acc-pr-%[1]s"
    billing_group = aiven_billing_group.foo.id
    parent_id     = aiven_organization.foo.id

    depends_on = [aiven_billing_group.foo]
}`, name)
}

func testAccBillingGroupResourceUpdated(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
    name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
    name                = "test-acc-bg-%[1]s"
    billing_emails      = ["ivan.savciuc+test2@aiven.fi", "ivan.savciuc+test3@aiven.fi"]
    billing_currency    = "EUR"
    billing_extra_text  = "updated reference number 456"
    company            = "Updated Company Ltd"
    address_lines      = ["456 Updated Street", "Suite 8", "Main Avenue"]
    country_code       = "FI"
    city              = "Updated City"
    zip_code          = "54321"
    state             = "Updated State"
    vat_id            = "UPDATED-VAT-456"
}

data "aiven_billing_group" "bar" {
    billing_group_id = aiven_billing_group.foo.id
}

resource "aiven_project" "pr1" {
    project       = "test-acc-pr-%[1]s"
    billing_group = aiven_billing_group.foo.id
    parent_id     = aiven_organization.foo.id

    depends_on = [aiven_billing_group.foo]
}`, name)
}

func testCopyBillingGroupFromExistingOne(name string) string {
	return fmt.Sprintf(`
resource "aiven_billing_group" "bar" {
  name             = "test-acc-bg-%s"
  billing_currency = "EUR"
  vat_id           = "abc"
  city             = "Helsinki"
  company          = "Aiven Oy"
}
resource "aiven_billing_group" "bar2" {
  name                    = "copy-test-acc-bg-%s"
  copy_from_billing_group = aiven_billing_group.bar.id
}`, name, name)
}
