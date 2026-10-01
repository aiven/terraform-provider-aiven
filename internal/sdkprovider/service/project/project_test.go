package project_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenProject_basic(t *testing.T) {
	resourceName := "aiven_project.foo"
	rName := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_cloud"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					// estimated_balance is deprecated and not populated on the resource, but available on the data source
					resource.TestCheckNoResourceAttr(resourceName, "estimated_balance"),
					resource.TestCheckNoResourceAttr(resourceName, "available_credits"),
					resource.TestCheckResourceAttrSet("data.aiven_project.project", "estimated_balance"),
					resource.TestCheckResourceAttrSet("data.aiven_project.project", "available_credits"),
				),
			},
			{
				Config:             testAccProjectDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

// TestAccAivenProject_copyFromProject tests copy_from_project with a billing group.
// Uses an existing aiven_organization_billing_group via AIVEN_BILLING_GROUP_ID.
// The variant that creates the billing group inline lives in
// TestAccAivenProject_broken because of backend restrictions documented there.
func TestAccAivenProject_copyFromProject(t *testing.T) {
	organizationID := acc.OrganizationID()
	billingGroupID := acc.BillingGroupID()

	resourceName := "aiven_project.foo"
	rName := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectCopyFromProjectResource(rName, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_cloud"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_group"),
				),
			},
		},
	})
}

// TestAccAivenProject_broken is a parking lot for subtests that currently cannot
// run against the backend. It is unconditionally skipped. Each subtest keeps the
// full fixture so that, when the underlying blockers are fixed, the scenario can
// be lifted back into a regular test by removing the surrounding t.Skip.
func TestAccAivenProject_broken(t *testing.T) {
	t.Skip("Parking lot for subtests blocked by backend restrictions; see each subtest comment.")

	organizationID := acc.OrganizationID()
	paymentMethodID := acc.PaymentMethodID()

	// copyFromProjectCreateBillingGroup is the "full fixture" variant of
	// TestAccAivenProject_copyFromProject: it creates the billing group inline from an
	// aiven_organization_address + the shared AIVEN_PAYMENT_METHOD_ID credit card, rather
	// than reusing an existing AIVEN_BILLING_GROUP_ID. Two things prevent it from running:
	//
	//  1. OrganizationBillingGroupCreate rejects the request with
	//     "[409] Credit card payment methods must use the same billing and shipping
	//     addresses across billing groups" because AIVEN_PAYMENT_METHOD_ID is already
	//     attached to a pre-existing billing group whose address records differ.
	//  2. Even when the billing group is created, the legacy /project create endpoint
	//     answers "[400] No billing information found. Please add the
	//     'use_source_project_billing_group' or 'billing_group_id' parameter" when the
	//     project's billing_group references an aiven_organization_billing_group id.
	//
	// Re-enable when either the backend accepts the id here or we have a dedicated
	// test credit card that is not already bound to another billing group.
	t.Run("copyFromProjectCreateBillingGroup", func(t *testing.T) {
		resourceName := "aiven_project.foo"
		rName := acc.RandStr()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenProjectResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccProjectCopyFromProjectCreateBillingGroupResource(rName, organizationID, paymentMethodID),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenProjectAttributes("data.aiven_project.project"),
						resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
						resource.TestCheckResourceAttrSet(resourceName, "default_cloud"),
						resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
						resource.TestCheckResourceAttrSet(resourceName, "billing_group"),
					),
				},
			},
		})
	})
}

func TestAccAivenProject_accounts(t *testing.T) {
	resourceName := "aiven_project.foo"
	rName := acc.RandStr()

	config := testAccProjectResourceAccounts(rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
				),
			},
			{
				// Tests account_id (deprecated) -> parent_id migration
				Config: regexp.MustCompile(`account_id\s+=`).ReplaceAllString(config, "parent_id ="),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "account_id"), // it is computed and should be set
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
				),
			},
		},
	})
}

func TestAccAivenProject_organizations(t *testing.T) {
	resourceName := "aiven_project.foo"
	rName := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceOrganizations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
				),
			},
		},
	})
}

func testAccProjectDoubleTagResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%s"
  account_id    = aiven_account.foo.account_id
  default_cloud = "aws-eu-west-2"
  tag {
    key   = "test"
    value = "val"
  }
  tag {
    key   = "test"
    value = "val2"
  }
}

data "aiven_project" "project" {
  project = aiven_project.foo.project

  depends_on = [aiven_project.foo]
}`, name, name)
}

func testAccProjectResourceAccounts(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%s"
  account_id    = aiven_account.foo.account_id
  default_cloud = "aws-eu-west-2"
  tag {
    key   = "test"
    value = "val"
  }
}

data "aiven_project" "project" {
  project = aiven_project.foo.project

  depends_on = [aiven_project.foo]
}`, name, name)
}

func testAccProjectResourceOrganizations(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%s"
  parent_id     = aiven_organization.foo.id
  default_cloud = "aws-eu-west-2"
  tag {
    key   = "test"
    value = "val"
  }
}`, name, name)
}

func testAccProjectResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "bar" {
  name = "test-acc-ac-%[1]s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%[1]s"
  account_id    = aiven_account.bar.account_id
  default_cloud = "aws-eu-west-2"
  tag {
    key   = "test"
    value = "val"
  }
}

data "aiven_project" "project" {
  project    = aiven_project.foo.project
  depends_on = [aiven_project.foo]
}`, name)
}

func testAccProjectCopyFromProjectResource(name, organizationID, billingGroupID string) string {
	return fmt.Sprintf(`
resource "aiven_project" "source" {
  project       = "test-acc-pr-source-%[1]s"
  parent_id     = %[2]q
  billing_group = %[3]q
  tag {
    key   = "test"
    value = "val"
  }
}

resource "aiven_project" "foo" {
  project           = "test-acc-pr-%[1]s"
  parent_id         = %[2]q
  billing_group     = %[3]q
  copy_from_project = aiven_project.source.project
}

data "aiven_project" "project" {
  project    = aiven_project.foo.project
  depends_on = [aiven_project.foo]
}`, name, organizationID, billingGroupID)
}

// testAccProjectCopyFromProjectCreateBillingGroupResource is the full-fixture variant
// used by the parked TestAccAivenProject_broken/copyFromProjectCreateBillingGroup
// subtest. It creates an aiven_organization_address + aiven_organization_billing_group
// in-test from the shared AIVEN_PAYMENT_METHOD_ID instead of reusing an existing
// AIVEN_BILLING_GROUP_ID. See the subtest comment for why this currently fails.
func testAccProjectCopyFromProjectCreateBillingGroupResource(name, organizationID, paymentMethodID string) string {
	return fmt.Sprintf(`
resource "aiven_organization_address" "addr" {
  organization_id = %[2]q
  address_lines   = ["123 Main St"]
  city            = "Helsinki"
  name            = "Test Company"
  country_code    = "FI"
  state           = "Uusimaa"
  zip_code        = "00100"
}

resource "aiven_organization_billing_group" "foo" {
  organization_id    = %[2]q
  billing_group_name = "test-acc-bg-%[1]s"
  billing_address_id = aiven_organization_address.addr.address_id
  billing_contact_emails {
    email = "contact@example.com"
  }
  billing_emails {
    email = "invoices@example.com"
  }
  payment_method {
    payment_method_id   = %[3]q
    payment_method_type = "credit_card"
  }
  shipping_address_id = aiven_organization_address.addr.address_id
  vat_id              = "123"
}

resource "aiven_project" "source" {
  project       = "test-acc-pr-source-%[1]s"
  parent_id     = %[2]q
  billing_group = aiven_organization_billing_group.foo.billing_group_id
  tag {
    key   = "test"
    value = "val"
  }
}

resource "aiven_project" "foo" {
  project                          = "test-acc-pr-%[1]s"
  parent_id                        = %[2]q
  copy_from_project                = aiven_project.source.project
  use_source_project_billing_group = true
}

data "aiven_project" "project" {
  project    = aiven_project.foo.project
  depends_on = [aiven_project.foo]
}`, name, organizationID, paymentMethodID)
}

func testAccCheckAivenProjectAttributes(n string, attributes ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] project attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["ca_cert"] == "" {
			return fmt.Errorf("expected to get an ca_cert from Aiven")
		}

		for _, attr := range attributes {
			if a[attr] == "" {
				return fmt.Errorf("expected to get an %s from Aiven", attr)
			}
		}

		return nil
	}
}

func testAccCheckAivenProjectResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project" {
			continue
		}

		p, err := c.Projects.Get(ctx, rs.Primary.ID)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 && e.Status != 403 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("project (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
