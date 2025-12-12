package unit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenOrganizationalUnit(t *testing.T) {
	var (
		resourceName = "aiven_organizational_unit.foo"
		orgName      = acc.OrganizationName()
		fooUnit      = acc.RandName("bar")
		barUnit      = acc.RandName("baz")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationalUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitResource(orgName, fooUnit, barUnit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fooUnit),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
					// Proves parent_id is org ID
					resource.TestCheckResourceAttrPair(
						"aiven_organizational_unit.foo", "parent_id",
						"data.aiven_organization.foo", "id",
					),
					resource.TestCheckResourceAttrWith("aiven_organizational_unit.foo", "parent_id", func(s string) error {
						if schemautil.IsOrganizationID(s) {
							return nil
						}
						return fmt.Errorf("expected parent_id to be an Organization ID, got %q", s)
					}),

					// Proves parent_id is account ID
					resource.TestCheckResourceAttrPair(
						"aiven_organizational_unit.bar", "parent_id",
						"data.aiven_account.bar", "account_id",
					),
					resource.TestCheckResourceAttrWith("aiven_organizational_unit.bar", "parent_id", func(s string) error {
						if schemautil.IsOrganizationID(s) {
							return fmt.Errorf("expected parent_id to be an Organization ID, got %q", s)
						}
						return nil
					}),
				),
			},
			{
				Config: testAccOrganizationalUnitResource(orgName, fooUnit+"_updated", barUnit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fooUnit+"_updated"),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccOrganizationalUnitResource(orgName, fooUnit, barUnit string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

data "aiven_account" "bar" {
  name = %[1]q
}

resource "aiven_organizational_unit" "foo" {
  parent_id = data.aiven_organization.foo.id
  name      = %[2]q
}

resource "aiven_organizational_unit" "bar" {
  parent_id = data.aiven_account.bar.account_id
  name      = %[3]q
}

// Finds data by name
data "aiven_organizational_unit" "foo" {
  name = aiven_organizational_unit.foo.name
}

// Finds data by id
data "aiven_organizational_unit" "bar" {
  id = aiven_organizational_unit.bar.id
}
`, orgName, fooUnit, barUnit)
}

func testAccCheckAivenOrganizationalUnitDestroy(s *terraform.State) error {
	var (
		c, err = acc.GetTestGenAivenClient()
		ctx    = context.Background()
	)

	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	// loop through the resources in state, verifying that organizational unit account is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organizational_unit" {
			continue
		}

		resp, err := c.AccountList(ctx)
		if err != nil {
			return fmt.Errorf("error listing accounts: %w", err)
		}

		for _, account := range resp {
			if account.AccountId == rs.Primary.ID {
				return fmt.Errorf("organizational unit (%q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
