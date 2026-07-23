package usergroup_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenOrganizationUserGroup_basic(t *testing.T) {
	resourceName := "aiven_organization_user_group.foo"
	datasourceName := "data.aiven_organization_user_group.bar"
	orgName := acc.OrganizationName()
	groupName := fmt.Sprintf("test-acc-u-grp-%s", acc.RandStr())
	updatedGroupName := fmt.Sprintf("test-acc-u-grp-%s", acc.RandStr())

	// Captured after creation to assert the group is updated in place
	// (same group_id) rather than recreated when name/description change.
	var groupID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationUserGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserGroupResource(orgName, groupName, "test"),
				Check: resource.ComposeTestCheckFunc(
					// Store the created group_id so a later step can assert it stays
					// the same, proving an in-place update over a recreation.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource %s not found in state", resourceName)
						}

						groupID = rs.Primary.Attributes["group_id"]

						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),

					// Datasource checks
					resource.TestCheckResourceAttrSet(datasourceName, "organization_id"),
					resource.TestCheckResourceAttr(datasourceName, "name", groupName),
					resource.TestCheckResourceAttr(datasourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(datasourceName, "group_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "create_time"),
					resource.TestCheckResourceAttrSet(datasourceName, "update_time"),
				),
			},
			{
				// Description is mutable, so updating it must not force recreation.
				Config: testAccOrganizationUserGroupResource(orgName, groupName, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
				),
			},
			{
				// name and description are both mutable, so updating them together
				// must update in place (same group_id) rather than force recreation.
				Config: testAccOrganizationUserGroupResource(orgName, updatedGroupName, "renamed description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "renamed description"),
					resource.TestCheckResourceAttrPtr(resourceName, "group_id", &groupID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAivenOrganizationUserGroup_backwardCompat verifies that state created by the
// previous SDK-based provider version is compatible with the Plugin Framework version.
func TestAccAivenOrganizationUserGroup_backwardCompat(t *testing.T) {
	resourceName := "aiven_organization_user_group.foo"
	orgName := acc.OrganizationName()
	groupName := fmt.Sprintf("test-acc-u-grp-%s", acc.RandStr())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           testAccOrganizationUserGroupResource(orgName, groupName, "test"),
			OldProviderVersion: "4.60.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "name", groupName),
				resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
				resource.TestCheckResourceAttrSet(resourceName, "group_id"),
				resource.TestCheckResourceAttr(resourceName, "description", "test"),
			),
		}),
	})
}

func testAccOrganizationUserGroupResource(orgName, groupName, description string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

resource "aiven_organization_user_group" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = %[2]q
  description     = %[3]q
}

data "aiven_organization_user_group" "bar" {
  organization_id = aiven_organization_user_group.foo.organization_id
  name            = aiven_organization_user_group.foo.name
}
`, orgName, groupName, description)
}

func testAccCheckAivenOrganizationUserGroupResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each organization user group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user_group" {
			continue
		}

		orgID, userGroupID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.UserGroupGet(ctx, orgID, userGroupID)
		if err != nil {
			if avngen.IsNotFound(err) {
				continue
			}

			return err
		}

		return fmt.Errorf("organization user group (%s) still exists", rs.Primary.ID)
	}

	return nil
}
