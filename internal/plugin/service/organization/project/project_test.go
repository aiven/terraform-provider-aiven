package project_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenOrganizationProjectResource = "aiven_organization_project"

// NEX-2895: creating aiven_organization_billing_group resources inside the test
// fails with
//   [409 OrganizationBillingGroupCreate]: Credit card payment methods must use
//   the same billing and shipping addresses across billing groups.
// because the shared AIVEN_PAYMENT_METHOD_ID credit card is already bound to a
// pre-existing billing group with different address records. Until the backend
// relaxes that or we have a dedicated test credit card, these tests reference an
// existing billing group via AIVEN_BILLING_GROUP_ID instead of creating one.

func TestAccAivenOrganizationProject(t *testing.T) {
	// Uses an existing aiven_organization_billing_group via AIVEN_BILLING_GROUP_ID
	// (see NEX-2895 note above).
	organizationID := acc.OrganizationID()
	billingGroupID := acc.BillingGroupID()

	resourceName := fmt.Sprintf("%s.foo", aivenOrganizationProjectResource)
	dataSourceName := fmt.Sprintf("data.%s.ds_test", aivenOrganizationProjectResource)
	rName := acc.RandStr()
	projectID := fmt.Sprintf("test-acc-org-pr-%s", rName)

	baseConfig := fmt.Sprintf(`
resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-unit-%[1]s"
  parent_id = %[2]q
}
`, rName, organizationID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationProjectResourceDestroy,
		Steps: []resource.TestStep{
			// test creating project with all possible fields
			{
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id = %[1]q

  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  technical_emails = ["john.doe+1@gmail.com", "john.doe+2@gmail.com"]

  tag {
    key   = "key1"
    value = "value1"
  }

  tag {
    key   = "key2"
    value = "value2"
  }

  tag {
    key   = "key3"
    value = "value3"
  }
}

data "aiven_organization_project" "ds_test" {
  project_id      = aiven_organization_project.foo.project_id
  organization_id = aiven_organization_project.foo.organization_id
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),

					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),

					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resourceName, "base_port"),

					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "2"), // Check number of emails
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+1@gmail.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+2@gmail.com"),

					resource.TestCheckResourceAttr(resourceName, "tag.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key3",
						"value": "value3",
					}),

					// test data source
					resource.TestCheckResourceAttrPair(dataSourceName, "project_id", resourceName, "project_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_id", resourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_group_id", resourceName, "billing_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "parent_id", resourceName, "parent_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "technical_emails.#", resourceName, "technical_emails.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tag.#", resourceName, "tag.#"),
				),
			},
			// test import state
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// test resource update
			{
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  # updating project_id without changing other billing_group_id would fail in this scenario
  project_id = %[1]q

  # should not change
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  technical_emails = ["john.doe+3@gmail.com", "john.doe+2@gmail.com", "john.doe+4@gmail.com"] #update emails

  tag { #update tags
    key   = "key1"
    value = "value1"
  }
  tag {
    key   = "key2"
    value = "value2-new"
  }
  tag {
    key   = "key4"
    value = "value4"
  }
}
`, projectID, organizationID, billingGroupID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						acc.ExpectOnlyAttributesChanged(resourceName, "technical_emails", "tag"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),

					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),

					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resourceName, "base_port"),

					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "3"), // Check number of emails
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+3@gmail.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+2@gmail.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+4@gmail.com"),

					resource.TestCheckResourceAttr(resourceName, "tag.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key2",
						"value": "value2-new",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key4",
						"value": "value4",
					}),
				),
			},
		},
	})
}

// TestAccAivenOrganizationProjectUpdateSteps tests the update steps of the aiven_organization_project resource.
//
// Note: this test used to exercise cross-organization scenarios (billing group in another
// organization, moving a project to another organization). These were removed when the
// deprecated aiven_billing_group was replaced with aiven_organization_billing_group:
// the latter requires a payment_method that is tied to one pre-existing organization,
// so a second organization cannot be stood up inside the test. The intra-organization
// scenarios below still provide coverage of create/update/import/replace flows.
func TestAccAivenOrganizationProjectUpdateSteps(t *testing.T) {
	// Uses an existing aiven_organization_billing_group via AIVEN_BILLING_GROUP_ID
	// (see NEX-2895 note above).
	organizationID := acc.OrganizationID()
	billingGroupID := acc.BillingGroupID()

	var (
		rName = acc.RandStr()

		resourceName     = fmt.Sprintf("%s.foo", aivenOrganizationProjectResource)
		dataSourceName   = fmt.Sprintf("data.%s.foo", aivenOrganizationProjectResource)
		projectID        = fmt.Sprintf("test-acc-pr-%s", rName)
		updatedProjectID = fmt.Sprintf("%s-new", projectID)

		// Basic configuration: two organizational units in the organization.
		// organization_id / billing_group_id are formatted directly into each step.
		baseConfig = generateBaseConfig(rName, organizationID)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				// basic creation with required fields without technical_emails and tags
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  base_port        = 10000
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "base_port", "10000"),
				),
			},
			{
				// test import state
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// updating with technical_emails
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  technical_emails = ["john.doe+1@gmail.com", "john.doe+2@gmail.com"]
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+1@gmail.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "technical_emails.*", "john.doe+2@gmail.com"),
				),
			},
			{
				// change parent_id which belongs to the same organization, should succeed. Also, remove technical_emails
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.fooz.id
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.fooz", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			{
				// Set parent_id to directly reference the organization ID instead of an organizational unit
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = %[2]q
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttr(resourceName, "parent_id", organizationID),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			// NEX-2895: the "change billing_group_id to a second billing group" step is
			// disabled until the test environment provides a second
			// AIVEN_BILLING_GROUP_ID. The backend forbids creating a second billing
			// group bound to the shared credit-card payment method with new address
			// records, so the test can only reference one existing billing group.
			// Original step kept here for reference:
			// {
			// 	// change billing_group_id which belongs to the same organization, should succeed
			// 	Config: baseConfig + fmt.Sprintf(`
			// resource "aiven_organization_project" "foo" {
			//   project_id       = %[1]q
			//   organization_id  = %[2]q
			//   billing_group_id = <second billing group id>
			//   parent_id        = aiven_organizational_unit.fooz.id
			// }
			// `, projectID, organizationID, billingGroupID),
			// 	Check: ...
			// },
			{
				// move parent_id back to an org unit (billing_group_id stays the same;
				// the "simultaneous billing group + parent_id" update is covered by the
				// disabled step above)
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			// Note: cross-organization "should fail" steps (billing group / parent_id in another
			// organization) and the "move project to another organization" step were removed when
			// aiven_billing_group was replaced with aiven_organization_billing_group. The new
			// resource requires a payment_method that is tied to one pre-existing organization,
			// so a second organization cannot be constructed inside the test.
			{
				// update project_id leads to new resource creation
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
}
`, updatedProjectID, organizationID, billingGroupID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", updatedProjectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
				),
			},
			{
				// rename project_id back leads to resource replacement
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID, organizationID, billingGroupID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
				),
			},
			{
				// update tags
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  tag {
    key   = "key1"
    value = "value1"
  }
  tag {
    key   = "key2"
    value = "value2"
  }
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
				),
			},
			{
				// Removes tags
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(resourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "0"),
				),
			},
			{
				// Brings tags back
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  tag {
    key   = "key1"
    value = "value1"
  }
  tag {
    key   = "key2"
    value = "value2"
  }
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tag.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
				),
			},
			{
				// test the data source
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = %[1]q
  organization_id  = %[2]q
  billing_group_id = %[3]q
  parent_id        = aiven_organizational_unit.foo.id
  tag {
    key   = "key1"
    value = "value1"
  }
  tag {
    key   = "key2"
    value = "value2"
  }
}

data "aiven_organization_project" "foo" {
  project_id      = aiven_organization_project.foo.project_id
  organization_id = aiven_organization_project.foo.organization_id
}
`, projectID, organizationID, billingGroupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(dataSourceName, "organization_id", organizationID),
					resource.TestCheckResourceAttr(dataSourceName, "billing_group_id", billingGroupID),
					resource.TestCheckResourceAttrPair(dataSourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ca_cert"),
					resource.TestCheckResourceAttr(dataSourceName, "tag.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "tag.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "tag.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
				),
			},
		},
	})
}

// generateBaseConfig creates two organizational units ("foo", "fooz") in the
// pre-existing organization (AIVEN_ORGANIZATION_ID). organization_id and
// billing_group_id are not exposed as locals: each step formats them directly
// into its own HCL via %[N]q, keeping the base config minimal.
//
// Billing groups are no longer created inside the test — see the NEX-2895 note
// at the top of the file. The second-billing-group scenario (switch billing
// groups on an existing project) is isolated and skipped below until a second
// AIVEN_BILLING_GROUP_ID can be provisioned.
//
// This also used to create a second organization ("bar") with its own billing
// group and organizational unit so cross-organization scenarios could be
// exercised; that was removed when aiven_billing_group was deprecated because
// aiven_organization_billing_group requires a payment_method tied to one
// pre-existing organization.
func generateBaseConfig(rName, organizationID string) string {
	return fmt.Sprintf(`
resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-unit-%[1]s"
  parent_id = %[2]q
}

resource "aiven_organizational_unit" "fooz" {
  name      = "test-acc-unit-%[1]s-fooz"
  parent_id = %[2]q
}
`, rName, organizationID)
}

func testAccCheckAivenOrganizationProjectResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error getting Aiven client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_project" {
			continue
		}

		orgID, projectID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing resource ID: %w", err)
		}

		resp, err := c.OrganizationProjectsList(ctx, orgID)
		if err != nil {
			if common.IsCritical(err) {
				return err
			}

			return nil // consider project as destroyed if it's not found
		}

		for _, p := range resp.Projects {
			if p.ProjectId == projectID {
				return fmt.Errorf("project (%q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
