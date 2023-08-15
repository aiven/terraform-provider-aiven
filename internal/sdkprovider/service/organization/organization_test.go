package organization_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganization_basic(t *testing.T) {
	t.Skip(
		"Skipping because aiven_organization is now implemented in the Terraform Plugin Framework version" +
			" of the provider, and this test is not yet ported to that framework.",
	)

	resourceName := "aiven_organizational_unit.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenOrganizationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenOrganizationUnitAttributes("data.aiven_organizational_unit.organizational_unit"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-org-unit-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "aiven"),
				),
			},
			{
				Config: testAccOrganizationUnitToProject(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aiven_project.pr", "account_id"),
				),
			},
			{
				Config: testAccOrganizationProjectDissociate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aiven_project.pr", "account_id", ""),
				),
			},
		},
	})
}

func testAccOrganizationResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-org-unit-%s"
  parent_id = data.aiven_organization.organization.id
}

data "aiven_organizational_unit" "organizational_unit" {
  name = aiven_organizational_unit.foo.name
}

data "aiven_organization" "organization" {
  name = aiven_organization.foo.name
}`, name, name)
}

func testAccOrganizationUnitToProject(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-org-unit-%s"
  parent_id = data.aiven_organization.organization.id
}

resource "aiven_project" "bar" {
  project    = "test-acc-org-unit-%s"
  account_id = aiven_organizational_unit.foo.id
}

data "aiven_project" "pr" {
  project = aiven_project.bar.project
}

data "aiven_organization" "organization" {
  name = aiven_organization.foo.name
}`, name, name, name)
}

func testAccOrganizationProjectDissociate(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_organizational_unit" "foo" {
  name      = "test-acc-org-unit-%s"
  parent_id = data.aiven_organization.organization.id
}

resource "aiven_project" "bar" {
  project = "test-acc-org-unit-%s"
}

data "aiven_project" "pr" {
  project = aiven_project.bar.project
}

data "aiven_organization" "organization" {
  name = aiven_organization.foo.name
}`, name, name, name)
}

func testAccCheckAivenOrganizationResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each organizational unit is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organizational_unit" {
			continue
		}

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, a := range r.Accounts {
			if a.Id == rs.Primary.ID {
				return fmt.Errorf("organizational unit (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAivenOrganizationUnitAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] organizational unit attributes %v", a)

		if a["parent_id"] == "" {
			return fmt.Errorf("expected to get a parent_id from Aiven")
		}

		if a["tenant_id"] == "" {
			return fmt.Errorf("expected to get a tenant_id from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["update_time"] == "" {
			return fmt.Errorf("expected to get a update_time from Aiven")
		}

		return nil
	}
}
