package billinggroup_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
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
				// Creates a group with billing_contact_emails
				Config: testAccBillingGroupResource(rName, `billing_contact_emails = ["foo@aiven.fi"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_contact_emails.#", "1"),
				),
			},
			{
				// Proves that billing_contact_emails can be removed (state update for nil value check)
				Config: testAccBillingGroupResource(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "billing_contact_emails"),
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
		return fmt.Errorf("error getting Aiven client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each billing group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_billing_group" {
			continue
		}

		db, err := c.BillingGroupGet(ctx, rs.Primary.ID)
		var e aiven.Error
		if common.IsCritical(err) && errors.As(err, &e) && e.Status != 500 {
			return fmt.Errorf("error getting a billing group by id: %w", err)
		}

		if db != nil {
			return fmt.Errorf("billing group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBillingGroupResource(name, contactEmails string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
  name           = "test-acc-bg-%[1]s"
  billing_emails = ["ivan.savciuc+test1@aiven.fi", "ivan.savciuc+test2@aiven.fi"]
  %s
}

data "aiven_billing_group" "bar" {
  billing_group_id = aiven_billing_group.foo.id
}

resource "aiven_project" "pr1" {
  project       = "test-acc-pr-%[1]s"
  billing_group = aiven_billing_group.foo.id
  parent_id     = aiven_organization.foo.id

  depends_on = [aiven_billing_group.foo]
}`, name, contactEmails)
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

func TestAccAivenBillingGroupDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_billing_group.bar"
	resourceName := "aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupResource(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckTypeSetElemAttr(datasourceName, "billing_emails.*", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.0"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.1"),
				),
			},
		},
	})
}

// TestAccAivenBillingGroup_backward_compatibility tests that resources created with the old SDK provider
// can be read by the new Plugin Framework provider.
// This is a regression test for the state migration issue where the old SDK state only had `id`
// but the new Plugin Framework expects `billing_group_id` attribute.
func TestAccAivenBillingGroup_backward_compatibility(t *testing.T) {
	resourceName := "aiven_billing_group.compat"
	rName := acc.RandName("compat-bg")
	orgName := acc.OrganizationName()

	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = "%[1]s"
}

resource "aiven_billing_group" "compat" {
  name             = "%[2]s"
  parent_id        = data.aiven_organization.org.id
  billing_currency = "EUR"
}
`, orgName, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckAivenBillingGroupResourceDestroy,
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.47.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(resourceName, "id"),
				resource.TestCheckResourceAttr(resourceName, "name", rName),
				resource.TestCheckResourceAttr(resourceName, "billing_currency", "EUR"),
			),
		}),
	})
}
