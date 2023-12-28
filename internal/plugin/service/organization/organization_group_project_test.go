package organization_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccOrganizationGroupProject(t *testing.T) {
	orgID, found := os.LookupEnv("AIVEN_ORG_ID")
	if !found {
		t.Skip("Skipping test due to missing AIVEN_ORG_ID environment variable")
	}

	if _, ok := os.LookupEnv("PROVIDER_AIVEN_ENABLE_BETA"); !ok {
		t.Skip("Skipping test due to missing PROVIDER_AIVEN_ENABLE_BETA environment variable")
	}

	suffix := acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)

	resourceName := "aiven_organization_group_project.foo"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserGroupProjectResource(suffix, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", "pr-"+suffix),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "role", "admin"),
				),
			},
		},
	})
}

func testAccOrganizationUserGroupProjectResource(rand, orgID string) string {
	return fmt.Sprintf(`
resource "aiven_organization_user_group" "foo" {
  organization_id = "%[1]s"
  name            = "test-group"
  description     = "test-group-description"
}

resource "aiven_project" "foo" {
  project   = "pr-%[2]s"
  parent_id = "%[1]s"
}
resource "aiven_organization_group_project" "foo" {
  project  = aiven_project.foo.project
  group_id = aiven_organization_user_group.foo.group_id
  role     = "admin"

  depends_on = [aiven_organization_user_group.foo, aiven_project.foo]
}
`, orgID, rand)
}
