package account_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenAccount_basic(t *testing.T) {
	resourceName := "aiven_account.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountResourceDestroy,
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
				Config: testAccAccountToProject(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aiven_project.pr", "account_id"),
				),
			},
			{
				Config: testAccAccountProjectDissociate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aiven_project.pr", "account_id", ""),
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

func testAccAccountProjectDissociate(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_project" "bar" {
  project = "test-acc-ac-%s"
}

data "aiven_project" "pr" {
  project = aiven_project.bar.project
}`, name, name)
}

func testAccCheckAivenAccountResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account" {
			continue
		}

		r, err := c.Accounts.List()
		if err != nil {
			var aivenError *aiven.Error

			if ok := errors.As(err, &aivenError); !ok || aivenError.Status != 404 {
				return err
			}

			return nil
		}

		for _, a := range r.Accounts {
			if a.Id == rs.Primary.ID {
				return fmt.Errorf("account (%s) still exists", rs.Primary.ID)
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

		return nil
	}
}
