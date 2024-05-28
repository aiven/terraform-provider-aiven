package organization_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccOrganizationApplicationUserDataSource(t *testing.T) {
	suffix := acc.RandStr()
	dataUserFoo := "data.aiven_organization_application_user.foo"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-org-app-user-%s"
}

data "aiven_organization_application_user" "foo" {
  organization_id = aiven_organization_application_user.foo.organization_id
  user_id         = aiven_organization_application_user.foo.user_id
}
`, os.Getenv("AIVEN_ORGANIZATION_NAME"), suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataUserFoo, "name", "test-acc-org-app-user-"+suffix),
					resource.TestCheckResourceAttr(dataUserFoo, "is_super_admin", "false"),
					resource.TestCheckResourceAttrSet(dataUserFoo, "email"),
				),
			},
		},
	})
}
