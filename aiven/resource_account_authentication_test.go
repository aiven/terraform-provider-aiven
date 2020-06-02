package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"testing"
)

func TestAccAivenAccountAuthentication_basic(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_account_authentication.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenAccountAuthenticationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAuthenticationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes("data.aiven_account_authentication.auth"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-auth-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
				),
			},
		},
	})
}

func testAccAccountAuthenticationResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_account" "foo" {
			name = "test-acc-ac-%s"
		}

		resource "aiven_account_authentication" "foo" {
  			account_id = aiven_account.foo.account_id
			name = "test-acc-auth-%s"
			type = "saml"
			enabled = false
		}

		data "aiven_account_authentication" "auth" {
  			account_id = aiven_account_authentication.foo.account_id
  			name = aiven_account_authentication.foo.name
		}
		`, name, name)
}

func testAccCheckAivenAccountAuthenticationResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account authentication is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_authentication" {
			continue
		}

		accountId, authId := splitResourceID2(rs.Primary.ID)

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, acc := range r.Accounts {
			if acc.Id == accountId {
				ra, err := c.AccountAuthentications.List(accountId)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
						return err
					}

					return nil
				}

				for _, a := range ra.AuthenticationMethods {
					if a.Id == authId {
						return fmt.Errorf("account authentication (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountAuthenticationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] account team attributes %v", a)

		if a["enabled"] != "false" {
			return fmt.Errorf("expected to get an enabled from Aiven")
		}

		if a["type"] != "saml" {
			return fmt.Errorf("expected to get a correty type from Aiven")
		}

		if a["account_id"] == "" {
			return fmt.Errorf("expected to get an account id from Aiven")
		}

		if a["authentication_id"] == "" {
			return fmt.Errorf("expected to get an authentication_id from Aiven")
		}

		if a["name"] == "" {
			return fmt.Errorf("expected to get a name from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["saml_acs_url"] == "" {
			return fmt.Errorf("expected to get a saml_acs_url from Aiven")
		}

		if a["saml_metadata_url"] == "" {
			return fmt.Errorf("expected to get a saml_metadata_url from Aiven")
		}

		return nil
	}
}
