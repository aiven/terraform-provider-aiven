package project_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenOrganizationProjectResource = "aiven_organization_project"

func TestAccAivenOrganizationProject(t *testing.T) {
	resourceName := fmt.Sprintf("%s.foo", aivenOrganizationProjectResource)
	dataSourceName := fmt.Sprintf("data.%s.ds_test", aivenOrganizationProjectResource)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	projectID := fmt.Sprintf("test-acc-org-pr-%s", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationProjectResourceDestroy,
		Steps: []resource.TestStep{
			// test creating project with all possible fields
			{
				Config: fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
  name      = "test-acc-bg-%[1]s"
  parent_id = aiven_organization.foo.id
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-unit-%[1]s"
  parent_id = aiven_organization.foo.id
}

resource "aiven_organization_project" "foo" {
  project_id = "%[2]s"

  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
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
`, rName, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),

					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"), // Check if organization_id is set to the correct value
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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
				Config: fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_billing_group" "foo" {
  name = "test-acc-bg-%[1]s"
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-unit-%[1]s"
  parent_id = aiven_organization.foo.id
}

resource "aiven_organization_project" "foo" {
  project_id = "%[2]s" #updating project_id without changing other billing_group_id would fail in this scenario

  organization_id  = aiven_organization.foo.id # should not change
  billing_group_id = aiven_billing_group.foo.id
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
`, rName,
					projectID,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						acc.ExpectOnlyAttributesChanged(resourceName, "technical_emails", "tag"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),

					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"), // Check if organization_id is set to the correct value
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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
func TestAccAivenOrganizationProjectUpdateSteps(t *testing.T) {
	acc.SkipIfNotBeta(t)

	var (
		rName = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resourceName     = fmt.Sprintf("%s.foo", aivenOrganizationProjectResource)
		dataSourceName   = fmt.Sprintf("data.%s.foo", aivenOrganizationProjectResource)
		projectID        = fmt.Sprintf("test-acc-pr-%s", rName)
		updatedProjectID = fmt.Sprintf("%s-new", projectID)

		// Basic configuration for organizations and their resources
		baseConfig = generateBaseConfig(rName)
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
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
  base_port        = 10000
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
  technical_emails = ["john.doe+1@gmail.com", "john.doe+2@gmail.com"]
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.fooz.id
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.fooz", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			{
				// Set parent_id to directly reference the organization ID instead of an organizational unit
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organization.foo.id
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			{
				// change billing_group_id which belongs to the same organization, should succeed
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.fooz.id
  parent_id        = aiven_organizational_unit.fooz.id
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.fooz", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.fooz", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			{
				// update billing group id and parent_id simultaneously
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "technical_emails.#", "0"),
				),
			},
			{
				// update billing group which belongs to a different organization, should fail
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.bar.id
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID),
				ExpectError: regexp.MustCompile(`Can't assign project to a billing group\s+belonging to a different org`),
			},
			{
				// update parent_id group which belongs to a different organization, should fail
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.bar.id
}
`, projectID),
				ExpectError: regexp.MustCompile(`Resources must belong to the same`),
			},
			{
				// update project_id leads to new resource creation
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
}
`, updatedProjectID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", updatedProjectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
				),
			},
			{
				// move project to another organization leads to new ID generation
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.bar.id
  billing_group_id = aiven_billing_group.bar.id
  parent_id        = aiven_organizational_unit.bar.id
}
`, updatedProjectID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", updatedProjectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.bar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.bar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.bar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
				),
			},
			{
				// move project to another organization with renaming simultaneously, leads to resource replacement
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
				),
			},
			{
				// update tags
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
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
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
  parent_id        = aiven_organizational_unit.foo.id
}
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "aiven_organizational_unit.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "0"),
				),
			},
			{
				// Brings tags back
				Config: baseConfig + fmt.Sprintf(`
resource "aiven_organization_project" "foo" {
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
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
`, projectID),
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
  project_id       = "%s"
  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id
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
`, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(dataSourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
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

// generateBaseConfig creates the base configuration for organizations, billing groups and organizational units
func generateBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_organization" "bar" {
  name = "test-acc-org-%[1]s-bar"
}

resource "aiven_billing_group" "foo" {
  name      = "test-acc-bg-%[1]s"
  parent_id = aiven_organization.foo.id
}

resource "aiven_billing_group" "fooz" {
  name      = "test-acc-bg-%[1]s-fooz"
  parent_id = aiven_organization.foo.id
}

resource "aiven_billing_group" "bar" {
  name      = "test-acc-bg-%[1]s-bar"
  parent_id = aiven_organization.bar.id
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-unit-%[1]s"
  parent_id = aiven_organization.foo.id
}

resource "aiven_organizational_unit" "fooz" {
  name      = "test-acc-unit-%[1]s-fooz"
  parent_id = aiven_organization.foo.id
}

resource "aiven_organizational_unit" "bar" {
  name      = "test-acc-unit-%[1]s-bar"
  parent_id = aiven_organization.bar.id
}
`, rName)
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
