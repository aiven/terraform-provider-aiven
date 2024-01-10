package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// TestAccOrganizationApplicationUserResourceDataSource tests the organization application user resource and data
// source.
func TestAccOrganizationApplicationUserResourceDataSource(t *testing.T) {
	deps := acc.CommonTestDependencies(t)

	deps.IsBeta(true)

	name := "aiven_organization_application_user.foo"
	dname := "data.aiven_organization_application_user.foo"

	suffix := acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
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
  name            = "%[1]s-org-appuser-%[2]s"
}
`, acc.DefaultResourceNamePrefix, suffix, deps.OrganizationName()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						name,
						"name",
						fmt.Sprintf("%s-org-appuser-%s", acc.DefaultResourceNamePrefix, suffix),
					),
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
						rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"],
					), nil
				},
			},
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[3]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%[1]s-org-appuser-%[2]s-1"
}
`, acc.DefaultResourceNamePrefix, suffix, deps.OrganizationName()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						name,
						"name",
						fmt.Sprintf("%s-org-appuser-%s-1", acc.DefaultResourceNamePrefix, suffix),
					),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[3]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%[1]s-org-appuser-%[2]s-1"
}

data "aiven_organization_application_user" "foo" {
  user_id         = aiven_organization_application_user.foo.user_id
  organization_id = data.aiven_organization.foo.id
}
`, acc.DefaultResourceNamePrefix, suffix, deps.OrganizationName()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						dname,
						"name",
						fmt.Sprintf("%s-org-appuser-%s-1", acc.DefaultResourceNamePrefix, suffix),
					),
				),
			},
		},
	})
}
