package organization_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func testAccOrganizationApplicationUserResource(suffix string, barAdmin bool) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[1]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-foo-%[2]s"
}

resource "aiven_organization_application_user" "bar" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-bar-%[2]s"
  is_super_admin  = %[3]t
}
`, os.Getenv("AIVEN_ORGANIZATION_NAME"), suffix, barAdmin)
}

func TestAccOrganizationApplicationUserResource(t *testing.T) {
	suffix := acc.RandStr()
	userFoo := "aiven_organization_application_user.foo"
	userBar := "aiven_organization_application_user.bar"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationApplicationUserResource(suffix, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Regular user
					resource.TestCheckResourceAttr(userFoo, "name", "test-acc-foo-"+suffix),
					resource.TestCheckResourceAttr(userFoo, "is_super_admin", "false"),
					resource.TestCheckResourceAttrSet(userFoo, "email"),

					// Admin user
					resource.TestCheckResourceAttr(userBar, "name", "test-acc-bar-"+suffix),
					resource.TestCheckResourceAttr(userBar, "is_super_admin", "true"),
					resource.TestCheckResourceAttrSet(userBar, "email"),
				),
			},
			{
				// Name admin is_super_admin update
				Config: testAccOrganizationApplicationUserResource(suffix, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Regular user
					resource.TestCheckResourceAttr(userFoo, "name", "test-acc-foo-"+suffix),
					resource.TestCheckResourceAttr(userFoo, "is_super_admin", "false"),
					resource.TestCheckResourceAttrSet(userFoo, "email"),

					// Admin user
					resource.TestCheckResourceAttr(userBar, "name", "test-acc-bar-"+suffix),
					resource.TestCheckResourceAttr(userBar, "is_super_admin", "false"),
					resource.TestCheckResourceAttrSet(userBar, "email"),
				),
			},
			{
				ResourceName:      userFoo,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, err := acc.ResourceFromState(state, userFoo)
					if err != nil {
						return "", err
					}

					return util.ComposeID(rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"]), nil
				},
			},
		},
	})
}
