package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAivenBillingGroup_basic(t *testing.T) {
	resourceName := "aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenBillingGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "billing_currency", "USD"),
				),
			},
		},
	})
}

func testAccCheckAivenBillingGroupResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each billing group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_billing_group" {
			continue
		}

		db, err := c.BillingGroup.Get(rs.Primary.ID)
		if err != nil && !aiven.IsNotFound(err) && err.(aiven.Error).Status != 500 {
			return fmt.Errorf("error getting a billing group by id: %w", err)
		}

		if db != nil {
			return fmt.Errorf("billing group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBillingGroupResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_billing_group" "foo" {
			name = "test-acc-bg-%s"
			billing_currency = "USD"
		}

		resource "aiven_project" "pr1" {
			project = "test-acc-pr-%s"
			billing_group = aiven_billing_group.foo.id

			depends_on = [aiven_billing_group.foo]
		}
		`, name, name)
}
