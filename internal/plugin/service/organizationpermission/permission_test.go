package organizationpermission_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// requireBillingGroupID returns the billing group for test projects: without one,
// the backend creates a default billing group that outlives the project.
func requireBillingGroupID(t *testing.T) string {
	billingGroupID := acc.BillingGroupID()
	if billingGroupID == "" {
		t.Skip("Skipping test due to missing AIVEN_BILLING_GROUP_ID environment variable")
	}
	return billingGroupID
}

func TestAccAivenOrganizationPermission_basic(t *testing.T) {
	billingGroupID := requireBillingGroupID(t)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	permission := "aiven_organization_permission.permission"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckBillingGroupsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPermissionResource(rName, billingGroupID, "developer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(permission, "resource_type", "project"),
					resource.TestCheckResourceAttr(permission, "permissions.#", "1"),
					resource.TestCheckResourceAttr(permission, "permissions.0.principal_type", "user_group"),
					resource.TestCheckResourceAttr(permission, "permissions.0.permissions.#", "1"),
					resource.TestCheckResourceAttr(permission, "permissions.0.permissions.0", "developer"),
				),
			},
			{
				Config: testAccOrganizationPermissionResource(rName, billingGroupID, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(permission, "permissions.0.permissions.0", "admin"),
				),
			},
			{
				// Removes permissions
				Config: testAccOrganizationPermissionResource(rName, billingGroupID, ""),
				Check: func(s *terraform.State) error {
					ctx := context.Background()
					client, err := acc.GetTestGenAivenClient()
					require.NoError(t, err)

					for _, r := range s.RootModule().Resources {
						if r.Type != "aiven_project" {
							continue
						}

						orgID := r.Primary.Attributes["parent_id"]
						permissions, err := client.PermissionsGet(ctx, orgID, organization.ResourceTypeProject, r.Primary.ID)
						if err != nil {
							return err
						}

						if len(permissions) > 0 {
							return fmt.Errorf("organization permissions (%s) still exists: %v", r.Primary.ID, permissions)
						}
					}
					return nil
				},
			},
		},
	})
}

func testAccOrganizationPermissionResource(name, billingGroupID, permission string) string {
	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %[1]q
}

resource "aiven_project" "project" {
  parent_id     = data.aiven_organization.org.id
  project       = "test-acc-permissions-%[2]s"
  billing_group = %[3]q
}

resource "aiven_organization_user_group" "group" {
  organization_id = data.aiven_organization.org.id
  name            = "test-acc-%[2]s"
  description     = "test group description"
}
`, acc.OrganizationName(), name, billingGroupID)

	if permission != "" {
		config += fmt.Sprintf(`
resource "aiven_organization_permission" "permission" {
  organization_id = data.aiven_organization.org.id
  resource_type   = "project"
  resource_id     = aiven_project.project.id

  permissions {
    principal_type = "user_group"
    principal_id   = aiven_organization_user_group.group.group_id
    permissions    = ["%s"]
  }
}
`, permission)
	}

	return config
}

func TestAccAivenOrganizationPermission_conflict(t *testing.T) {
	billingGroupID := requireBillingGroupID(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckBillingGroupsDestroyed,
		Steps: []resource.TestStep{
			{
				ExpectError: regexp.MustCompile("already has permissions configured"),
				Config: fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %[1]q
}

resource "aiven_project" "project" {
  parent_id     = data.aiven_organization.org.id
  project       = "test-acc-%[2]s"
  billing_group = %[3]q
}

resource "aiven_organization_user_group" "group" {
  organization_id = data.aiven_organization.org.id
  name            = "test-acc-%[2]s"
  description     = "test group description"
}

resource "aiven_organization_permission" "first" {
  organization_id = data.aiven_organization.org.id
  resource_type   = "project"
  resource_id     = aiven_project.project.id

  permissions {
    principal_type = "user_group"
    principal_id   = aiven_organization_user_group.group.group_id
    permissions    = ["developer"]
  }
}

resource "aiven_organization_permission" "second_conflicting" {
  organization_id = data.aiven_organization.org.id
  resource_type   = "project"
  resource_id     = aiven_project.project.id

  permissions {
    principal_type = "user_group"
    principal_id   = aiven_organization_user_group.group.group_id
    permissions    = ["developer"]
  }
}
`, acc.OrganizationName(), acc.RandStr(), billingGroupID),
			},
		},
	})
}
