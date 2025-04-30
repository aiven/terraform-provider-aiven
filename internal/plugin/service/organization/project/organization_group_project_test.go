package project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// TestAccOrganizationGroupProject tests the organization group project relation resource.
func TestAccOrganizationGroupProject(t *testing.T) {
	t.Skip("Deprecated resource")

	acc.SkipIfNotBeta(t)

	name := "aiven_organization_group_project.foo"

	suffix := acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[3]s"
}

resource "aiven_organization_user_group" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%[1]s-usr-group-%[2]s"
  description     = "Terraform acceptance tests"
}

resource "aiven_project" "foo" {
  project   = "%[1]s-pr-%[2]s"
  parent_id = data.aiven_organization.foo.id
}

resource "aiven_organization_group_project" "foo" {
  project  = aiven_project.foo.project
  group_id = aiven_organization_user_group.foo.group_id
  role     = "admin"
}
`, acc.DefaultResourceNamePrefix, suffix, acc.OrganizationName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						name,
						"project",
						fmt.Sprintf("%s-pr-%s", acc.DefaultResourceNamePrefix, suffix),
					),
					resource.TestCheckResourceAttrSet(name, "group_id"),
					resource.TestCheckResourceAttr(name, "role", "admin"),
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

					return util.ComposeID(rs.Primary.Attributes["project"], rs.Primary.Attributes["group_id"]), nil
				},
			},
		},
	})
}
