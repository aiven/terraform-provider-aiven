package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func testAccOrganizationApplicationUserResource(suffix string, isSuperAdmin bool) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[1]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-foo-%[2]s"
  is_super_admin  = %t
}
`, acc.OrganizationName(), suffix, isSuperAdmin)
}

func TestAccOrganizationApplicationUserResource(t *testing.T) {
	suffix := acc.RandStr()
	resourceName := "aiven_organization_application_user.foo"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationApplicationUserResource(suffix, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-foo-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "is_super_admin", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "email"),
				),
			},
			{
				Config:             testAccOrganizationApplicationUserResource(suffix, true),
				PlanOnly:           false,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The value can't be changed and remains false.
					resource.TestCheckResourceAttr(resourceName, "is_super_admin", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, err := acc.ResourceFromState(state, resourceName)
					if err != nil {
						return "", err
					}

					return util.ComposeID(rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"]), nil
				},
			},
		},
	})
}
