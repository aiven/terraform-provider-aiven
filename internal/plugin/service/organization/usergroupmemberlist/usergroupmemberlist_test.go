package usergroupmemberlist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccOrganizationUserGroupListMember(t *testing.T) {
	acc.SkipIfNotBeta(t)

	name := "data.aiven_organization_user_group_member_list.foo"
	userName := acc.RandName("user")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

resource "aiven_organization_user_group" "foo" {
  name            = %[2]q
  organization_id = data.aiven_organization.foo.id
  description     = "Terraform acceptance tests"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = %[3]q
}

resource "aiven_organization_user_group_member" "foo" {
  organization_id = data.aiven_organization.foo.id
  group_id        = aiven_organization_user_group.foo.group_id
  user_id         = aiven_organization_application_user.foo.user_id
}

data "aiven_organization_user_group_member_list" "foo" {
  organization_id = data.aiven_organization.foo.id
  user_group_id   = aiven_organization_user_group.foo.group_id

  depends_on = [
    aiven_organization_user_group_member.foo,
  ]
}
`, acc.OrganizationName(), acc.RandName("group"), userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "members.#", "1"),
					resource.TestCheckResourceAttr(name, "members.0.user_info.0.real_name", userName),
					resource.TestCheckResourceAttr(name, "members.0.user_info.0.is_application_user", "true"),
				),
			},
		},
	})
}
