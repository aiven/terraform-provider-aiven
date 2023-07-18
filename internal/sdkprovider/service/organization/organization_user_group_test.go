package organization_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenOrganizationUserGroup_basic(t *testing.T) {
	resourceName := "aiven_organization_user_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenOrganizationUserGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenOrganizationUserGroupAttributes("data.aiven_organization_user_group.bar"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-u-grp-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccOrganizationUserGroupResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_organization_user_group" "foo" {
  name            = "test-acc-u-grp-%s"
  organization_id = aiven_organization.foo.id
  description     = "test"
}

data "aiven_organization_user_group" "bar" {
  name            = aiven_organization_user_group.foo.name
  organization_id = aiven_organization_user_group.foo.organization_id
}
`, name, name)
}

func testAccCheckAivenOrganizationUserGroupResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each organization user group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user_group" {
			continue
		}

		orgID, userGroupID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.OrganizationUserGroups.Get(orgID, userGroupID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		if r != nil {
			return fmt.Errorf("organization user group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAivenOrganizationUserGroupAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] organization user group attributes %v", a)

		if a["organization_id"] == "" {
			return fmt.Errorf("expected to get a organization_id from Aiven")
		}

		if a["name"] == "" {
			return fmt.Errorf("expected to get a name from Aiven")
		}

		if a["description"] == "" {
			return fmt.Errorf("expected to get a description from Aiven")
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
