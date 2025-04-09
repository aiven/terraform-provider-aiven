package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// TestAccOrganizationUserGroupMember tests the organization user group member resource.
func TestAccOrganizationUserGroupMember(t *testing.T) {
	acc.SkipIfNotBeta(t)

	name := "aiven_organization_user_group_member.foo"
	suffix := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[3]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "foo"
}

resource "aiven_organization_user_group" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%[1]s-usr-group-%[2]s"
  description     = "Terraform acceptance tests"
}

resource "aiven_organization_user_group_member" "foo" {
  organization_id = data.aiven_organization.foo.id
  group_id        = aiven_organization_user_group.foo.group_id
  user_id         = aiven_organization_application_user.foo.user_id
}
	`, acc.DefaultResourceNamePrefix, suffix, acc.OrganizationName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, err := acc.ResourceFromState(state, name)
					if err != nil {
						return "", err
					}

					return util.ComposeID(
						rs.Primary.Attributes["organization_id"],
						rs.Primary.Attributes["group_id"],
						rs.Primary.Attributes["user_id"],
					), nil
				},
			},
		},
	})
}
