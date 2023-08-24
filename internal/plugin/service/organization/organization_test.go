package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccOrganizationResourceDataSource tests the organization resource and data source.
func TestAccOrganizationResourceDataSource(t *testing.T) {
	name := "aiven_organization.foo"

	dnames := []string{
		"data.aiven_organization.foo",
		"data.aiven_organization.bar",
	}

	suffix := acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "%s-org-%s"
}
`, acc.DefaultResourceNamePrefix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "name", fmt.Sprintf("%s-org-%s", acc.DefaultResourceNamePrefix, suffix),
					),
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttrSet(name, "tenant_id"),
					resource.TestCheckResourceAttrSet(name, "create_time"),
					resource.TestCheckResourceAttrSet(name, "update_time"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "%s-org-%s-1"
}
`, acc.DefaultResourceNamePrefix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "name", fmt.Sprintf("%s-org-%s-1", acc.DefaultResourceNamePrefix, suffix),
					),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "%s-org-%s-1"
}

data "aiven_organization" "foo" {
  id = aiven_organization.foo.id
}

data "aiven_organization" "bar" {
  name = aiven_organization.foo.name
}
`, acc.DefaultResourceNamePrefix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						dnames[0], "name", fmt.Sprintf("%s-org-%s-1", acc.DefaultResourceNamePrefix, suffix),
					),
					resource.TestCheckResourceAttrPair(
						dnames[1], "id", name, "id",
					),
				),
			},
		},
	})
}
