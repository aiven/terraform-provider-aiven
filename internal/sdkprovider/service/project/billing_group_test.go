package project_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenBillingGroup_basic(t *testing.T) {
	resourceName := "aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenBillingGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-bg-%s", rName)),
				),
			},
			{
				Config: testCopyBillingGroupFromExistingOne(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "name", fmt.Sprintf("copy-test-acc-bg-%s", rName)),
					resource.TestCheckResourceAttr("aiven_billing_group.bar", "billing_currency", "EUR"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "billing_currency", "EUR"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "city", "Helsinki"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "company", "Aiven Oy"),
					resource.TestCheckResourceAttr("aiven_billing_group.bar2", "vat_id", "abc"),
				),
			},
		},
	})
}

func testAccCheckAivenBillingGroupResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each billing group is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_billing_group" {
			continue
		}

		db, err := c.BillingGroup.Get(ctx, rs.Primary.ID)
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
  name           = "test-acc-bg-%s"
  billing_emails = ["ivan.savciuc+test1@aiven.fi", "ivan.savciuc+test2@aiven.fi"]
}

data "aiven_billing_group" "bar" {
  billing_group_id = aiven_billing_group.foo.id
}

resource "aiven_project" "pr1" {
  project       = "test-acc-pr-%s"
  billing_group = aiven_billing_group.foo.id

  depends_on = [aiven_billing_group.foo]
}`, name, name)
}

func testCopyBillingGroupFromExistingOne(name string) string {
	return fmt.Sprintf(`
resource "aiven_billing_group" "bar" {
  name             = "test-acc-bg-%s"
  billing_currency = "EUR"
  vat_id           = "abc"
  city             = "Helsinki"
  company          = "Aiven Oy"
}
resource "aiven_billing_group" "bar2" {
  name                    = "copy-test-acc-bg-%s"
  copy_from_billing_group = aiven_billing_group.bar.id
}`, name, name)
}
