package account_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
)

func TestAccAivenAccount_basic(t *testing.T) {
	resourceName := "aiven_account.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAttributes("data.aiven_account.account"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-ac-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "aiven"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_billing_group_id"),
				),
			},
			{
				// change the account name and check that it will be updated
				Config:             testAccAccountResource(fmt.Sprintf("%s-new", rName)),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAccountToProject(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aiven_project.pr", "account_id"),
				),
			},
		},
	})
}

func testAccAccountResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_billing_group" "bar" {
  name = "test-acc-bg-%s"
}

resource "aiven_account" "foo" {
  name                     = "test-acc-ac-%s"
  primary_billing_group_id = aiven_billing_group.bar.id
}

data "aiven_account" "account" {
  name = aiven_account.foo.name
}`, name, name)
}

func testAccAccountToProject(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_project" "bar" {
  project    = "test-acc-ac-%s"
  account_id = aiven_account.foo.account_id
}

data "aiven_project" "pr" {
  project = aiven_project.bar.project
}`, name, name)
}

func testAccCheckAivenAccountResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error instantiating client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each account is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account" {
			continue
		}

		resp, err := c.AccountList(ctx)
		if common.IsCritical(err) {
			return err
		}

		for _, account := range resp {
			if account.AccountId == rs.Primary.ID {
				return fmt.Errorf("account (%q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] account attributes %v", a)

		if a["account_id"] == "" {
			return fmt.Errorf("expected to get an account id from Aiven")
		}

		if a["owner_team_id"] == "" {
			return fmt.Errorf("expected to get a owner_team_id from Aiven")
		}

		if a["tenant_id"] == "" {
			return fmt.Errorf("expected to get a tenant_id from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["update_time"] == "" {
			return fmt.Errorf("expected to get a update_time from Aiven")
		}

		if a["is_account_owner"] == "" {
			return fmt.Errorf("expected to get a is_account_owner from Aiven")
		}

		return nil
	}
}
