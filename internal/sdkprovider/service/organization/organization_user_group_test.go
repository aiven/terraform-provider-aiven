package organization_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/usergroup"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenOrganizationUserGroup_basic(t *testing.T) {
	resourceName := "aiven_organization_user_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationUserGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenOrganizationUserGroupAttributes("data.aiven_organization_user_group.bar"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-u-grp-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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
	var (
		c, err = acc.GetTestGenAivenClient()
		ctx    = context.Background()
	)

	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	// loop through the resources in state, verifying each organization user group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user_group" {
			continue
		}

		orgID, userGroupID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.UserGroupGet(ctx, orgID, userGroupID)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
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

func TestAccAivenOrganizationUserGroup_Import(t *testing.T) {
	var (
		resourceName = "aiven_organization_user_group.import_group"
		rName        = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		orgID        string
		groupID      string
	)

	acc.TestAccPreCheck(t)

	orgID = os.Getenv("AIVEN_ORG_ID")
	if orgID == "" {
		t.Skip("Skipping test due to missing AIVEN_ORG_ID environment variable")
	}

	// create organization and group before test
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		t.Fatalf("failed to get client: %s", err)
	}

	userGroup, err := c.UserGroupCreate(context.Background(), orgID, &usergroup.UserGroupCreateIn{
		Description:   "Imported group",
		UserGroupName: fmt.Sprintf("test-acc-import-group-%s", rName),
	})
	if err != nil {
		t.Fatalf("failed to create user group: %s", err)
	}
	groupID = userGroup.UserGroupId

	// cleanup the user group in case of failure
	t.Cleanup(func() {
		if err = c.UserGroupDelete(context.Background(), orgID, groupID); common.IsCritical(err) {
			t.Errorf("failed to delete user group: %s", err)
		}
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationUserGroupResourceDestroy,
		Steps: []resource.TestStep{
			// first verify that plan is empty
			{
				Config:             testAccOrganizationUserGroupImportConfig(orgID, fmt.Sprintf("test-acc-import-group-%s", rName), groupID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// then actually perform the import and verify attributes
			{
				Config: testAccOrganizationUserGroupImportConfig(orgID, fmt.Sprintf("test-acc-import-group-%s", rName), groupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-import-group-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Imported group"),
					resource.TestCheckResourceAttr(resourceName, "organization_id", orgID),
					resource.TestCheckResourceAttr(resourceName, "group_id", groupID),
				),
			},
		},
	})
}

func testAccOrganizationUserGroupImportConfig(orgID, name, groupID string) string {
	return fmt.Sprintf(`
resource "aiven_organization_user_group" "import_group" {
  organization_id = "%s"
  name            = "%s"
  description     = "Imported group"
}

import {
  id = "%s/%s"
  to = aiven_organization_user_group.import_group
}
`, orgID, name, orgID, groupID)
}
