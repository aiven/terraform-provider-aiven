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
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/project"
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
  name = "test-acc-bg-%[1]s"
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

func TestAccAivenOrganizationProjectUpdateStages(t *testing.T) {
	var (
		registry       = acc.NewTemplateRegistry(aivenOrganizationProjectResource)
		rName          = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		newComposition = func() *acc.CompositionBuilder {
			return registry.NewCompositionBuilder()
		}

		resourceName     = fmt.Sprintf("%s.foo", aivenOrganizationProjectResource)
		projectID        = fmt.Sprintf("test-acc-pr-%s", rName)
		updatedProjectID = fmt.Sprintf("%s-new", projectID)
	)

	preSetOrganizationProjectTemplates(t, registry)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				// basic creation with required fields without parent_id
				Config: newComposition().
					Add("project", map[string]interface{}{
						"name":       rName,
						"project_id": projectID,
						"ref_org":    "foo",
						"ref_bg":     "foo",
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					// parent_id would be set even if not provided. It would be set to the organization account ID
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
				),
			},
			{
				// test import state
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// try to change only parent_id - should fail
				Config: newComposition().
					Add("project", map[string]interface{}{
						"name":       rName,
						"project_id": projectID,
						"ref_org":    "foo",
						"ref_bg":     "foo",
						"ref_par":    "bar", // trying to change parent_id group only
					}).MustRender(t),
				ExpectError: regexp.MustCompile(project.ErrProjectStructureChangeNotSupported.Error()),
			},
			{
				// try to change only billing_group_id - should fail
				Config: newComposition().
					Add("project", map[string]interface{}{
						"name":       rName,
						"project_id": projectID,
						"ref_org":    "foo",
						"ref_bg":     "bar", // trying to change billing group only
						"ref_par":    "foo",
					}).MustRender(t),
				ExpectError: regexp.MustCompile(project.ErrProjectStructureChangeNotSupported.Error()),
			},
			{
				// update project_id
				Config: newComposition().
					Add("project", map[string]interface{}{
						"name":       rName,
						"project_id": updatedProjectID,
						"ref_org":    "foo",
						"ref_bg":     "foo",
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", updatedProjectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.foo", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.foo", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
				),
			},
			{
				// update all
				Config: newComposition().
					Add("project", map[string]interface{}{
						"name":       rName,
						"project_id": updatedProjectID,
						"ref_org":    "bar",
						"ref_bg":     "bar",
						"ref_par":    "bar",
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", updatedProjectID),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "aiven_organization.bar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "billing_group_id", "aiven_billing_group.bar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
				),
			},
		},
	})
}

func preSetOrganizationProjectTemplates(t *testing.T, registry *acc.TemplateRegistry) {
	registry.MustAddTemplate(t, "project", `
resource "aiven_organization" "foo" {
    name = "test-acc-org-{{.name}}"
}

resource "aiven_organization" "bar" {
    name = "test-acc-org-{{.name}}-bar"
}

resource "aiven_billing_group" "foo" {
    name = "test-acc-bg-{{.name}}"
	parent_id = aiven_organization.foo.id
}

resource "aiven_billing_group" "bar" {
    name = "test-acc-bg-{{.name}}-bar"
	parent_id = aiven_organization.bar.id
}

resource "aiven_organizational_unit" "foo" {
    name      = "test-acc-unit-{{.name}}"
    parent_id = aiven_organization.foo.id
}

resource "aiven_organizational_unit" "bar" {
    name      = "test-acc-unit-{{.name}}-bar"
    parent_id = aiven_organization.bar.id
}

resource "aiven_organization_project" "foo" {
    project_id       = "{{.project_id}}"
    organization_id  = aiven_organization.{{.ref_org}}.id
    billing_group_id = aiven_billing_group.{{.ref_bg}}.id
	{{- if .ref_par}}
    parent_id        = aiven_organizational_unit.{{.ref_par}}.id
    {{- end}}

    {{- if .technical_emails}}
    technical_emails = [{{range .technical_emails}}"{{.}}",{{end}}]
    {{- end}}

    {{- if .tags}}
    {{- range .tags}}
    tag {
        key   = "{{.key}}"
        value = "{{.value}}"
    }
    {{- end}}
    {{- end}}
}`)
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

			return nil //consider project as destroyed if it's not found
		}

		for _, p := range resp.Projects {
			if p.ProjectId == projectID {
				return fmt.Errorf("project (%q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
