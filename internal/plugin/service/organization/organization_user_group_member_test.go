package organization_test

import (
	"fmt"
	"os"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationUserGroupMemeber(t *testing.T) {
	orgID := os.Getenv("AIVEN_ORG_ID")
	userID := os.Getenv("AIVEN_ORG_USER_ID")

	if orgID == "" || userID == "" {
		t.Skip("Skipping test due to missing AIVEN_ORG_ID or AIVEN_ORG_USER_ID environment variable")
	}

	if os.Getenv("PROVIDER_AIVEN_ENABLE_BETA") == "" {
		t.Skip("Skipping test due to missing PROVIDERX_ENABLE_BETA environment variable")
	}

	resourceName := "aiven_organization_user_group_member.foo"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserGroupMemberResource(orgID, userID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "last_activity_time"),
				),
			},
		},
	})
}

func testAccOrganizationUserGroupMemberResource(orgID, userID string) string {
	return fmt.Sprintf(`
resource "aiven_organization_user_group" "foo" {
  organization_id = "%[1]s"
  name            = "testacc-dummy-user-group"
  description     = "testacc-dummy-user-group-description"
}

resource "aiven_organization_user_group_member" "foo" {
  organization_id = "%[1]s"
  group_id        = aiven_organization_user_group.foo.group_id
  user_id         = "%[2]s"
}
	`, orgID, userID)
}
