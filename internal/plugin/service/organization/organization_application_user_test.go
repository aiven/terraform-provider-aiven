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

	names := []string{
		"aiven_organization_application_user.foo",
		"aiven_organization_application_user_token.foo",
	}

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
						names[0],
						"name",
						fmt.Sprintf("%s-org-appuser-%s", acc.DefaultResourceNamePrefix, suffix),
					),
					resource.TestCheckResourceAttrSet(names[0], "id"),
				),
			},
			{
				ResourceName:      names[0],
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, err := acc.ResourceFromState(state, names[0])
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
						names[0],
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
  organization_id = data.aiven_organization.foo.id
  user_id         = aiven_organization_application_user.foo.user_id
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
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[3]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%[1]s-org-appuser-%[2]s-1"
}

resource "aiven_organization_application_user_token" "foo" {
  organization_id  = aiven_organization_application_user.foo.organization_id
  user_id          = aiven_organization_application_user.foo.user_id
  description      = "Terraform acceptance tests"
  max_age_seconds  = 3600
  extend_when_used = true
  scopes           = ["user:read"]
}
`, acc.DefaultResourceNamePrefix, suffix, deps.OrganizationName()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(names[1], "description", "Terraform acceptance tests"),
					resource.TestCheckResourceAttr(names[1], "max_age_seconds", "3600"),
					resource.TestCheckResourceAttr(names[1], "extend_when_used", "true"),
					resource.TestCheckResourceAttr(names[1], "scopes.#", "1"),
				),
			},
		},
	})
}
