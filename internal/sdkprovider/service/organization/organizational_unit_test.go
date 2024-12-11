package organization_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

const orgUnitResource = "aiven_organizational_unit"

func TestAccAivenOrganizationalUnit(t *testing.T) {
	var (
		resourceName = fmt.Sprintf(orgUnitResource + ".bar")
		rName        = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationalUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-unit-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
			{
				Config: testAccOrganizationalUnitResource(rName + "_updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-unit-%s", rName+"_updated")),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccOrganizationalUnitResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-orgu-%s"
}

resource "aiven_organizational_unit" "bar" {
  name      = "test-acc-unit-%s"
  parent_id = aiven_organization.foo.id
}
`, name, name)
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
		if rs.Type != orgUnitResource {
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
