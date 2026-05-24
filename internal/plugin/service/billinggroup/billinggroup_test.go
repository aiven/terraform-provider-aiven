package billinggroup_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
)

// TestAccAivenBillingGroup_basic creates a dedicated organization to avoid hitting
// the backend limit of 5 billing groups per organization when running in parallel CI.
func TestAccAivenBillingGroup_basic(t *testing.T) {
	resourceName := "aiven_billing_group.foo"
	datasourceName := "data.aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenBillingGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupBasicResource(rName, `billing_contact_emails = ["foo@aiven.fi"]`),
				Check: resource.ComposeTestCheckFunc(
					// Creates a group with billing_contact_emails
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_contact_emails.#", "1"),

					// Datasource test
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckTypeSetElemAttr(datasourceName, "billing_emails.*", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.0"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.1"),
				),
			},
			{
				// Proves that billing_contact_emails can be removed (state update for nil value check)
				Config: testAccBillingGroupBasicResource(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "billing_contact_emails"),
				),
			},
			{
				Config: testAccBillingGroupBasicResource(rName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: testAccBillingGroupBasicResource(rName, `billing_contact_emails = []`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "billing_contact_emails.#", "0"),
				),
			},
			{
				Config: testAccBillingGroupBasicResource(rName, `billing_contact_emails = []`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccAivenBillingGroup_clone(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenBillingGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupCloneResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aiven_billing_group.clone", "name", fmt.Sprintf("test-acc-bg-copy-%s", rName)),
					resource.TestCheckResourceAttr("aiven_billing_group.clone", "billing_currency", "EUR"),
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

func testAccBillingGroupBasicResource(rName, contactEmails string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "org" {
  name = "test-acc-bg-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
  parent_id      = aiven_organization.org.id
  name           = "test-acc-bg-%[1]s"
  billing_emails = ["ivan.savciuc+test1@aiven.fi", "ivan.savciuc+test2@aiven.fi"]
  %[2]s
}

data "aiven_billing_group" "foo" {
  billing_group_id = aiven_billing_group.foo.id
}
`, rName, contactEmails)
}

func testAccBillingGroupCloneResource(rName string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "org" {
  name = "test-acc-bg-clone-org-%[1]s"
}

// A source billing group to copy from without emails
// that can cause plan diff issues and fail the test
resource "aiven_billing_group" "source" {
  parent_id        = aiven_organization.org.id
  name             = "test-acc-bg-source-%[1]s"
  billing_currency = "EUR"
}

resource "aiven_billing_group" "clone" {
  name                    = "test-acc-bg-copy-%[1]s"
  parent_id               = aiven_organization.org.id
  copy_from_billing_group = aiven_billing_group.source.id
}
`, rName)
}
