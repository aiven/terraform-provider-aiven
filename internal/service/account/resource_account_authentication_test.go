package account_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenAccountAuthentication_basic(t *testing.T) {
	resourceName := "aiven_account_authentication.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountAuthenticationResourceDestroy,
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

func TestAccAivenAccountAuthentication_auto_join_team_id(t *testing.T) {
	resourceName := "aiven_account_authentication.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountAuthenticationWithAutoJoinTeamIDResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAuthenticationWithAutoJoinTeamIDResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes("data.aiven_account_authentication.auth"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-auth-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_join_team_id", "aiven_account_team.foo", "team_id"),
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
  name       = "test-acc-auth-%s"
  type       = "saml"
  enabled    = false
}

data "aiven_account_authentication" "auth" {
  account_id = aiven_account_authentication.foo.account_id
  name       = aiven_account_authentication.foo.name

  depends_on = [aiven_account_authentication.foo]
}`, name, name)
}

func testAccAccountAuthenticationWithAutoJoinTeamIDResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name       = "test-acc-team-%s"
}

resource "aiven_account_authentication" "foo" {
  account_id        = aiven_account.foo.account_id
  name              = "test-acc-auth-%s"
  type              = "saml"
  enabled           = false
  auto_join_team_id = aiven_account_team.foo.team_id
}

data "aiven_account_team" "team" {
  name       = aiven_account_team.foo.name
  account_id = aiven_account_team.foo.account_id

  depends_on = [aiven_account_team.foo]
}

data "aiven_account_authentication" "auth" {
  account_id = aiven_account_authentication.foo.account_id
  name       = aiven_account_authentication.foo.name

  depends_on = [aiven_account_authentication.foo]
}`, name, name, name)
}

func testAccCheckAivenAccountAuthenticationResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account authentication is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_authentication" {
			continue
		}

		accountId, authId := schemautil.SplitResourceID2(rs.Primary.ID)

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, ac := range r.Accounts {
			if ac.Id == accountId {
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

func testAccCheckAivenAccountAuthenticationWithAutoJoinTeamIDResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account authentication is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team" && rs.Type != "aiven_account_authentication" {
			continue
		}

		isTeam := rs.Type == "aiven_account_team"

		accountID, secondaryID := schemautil.SplitResourceID2(rs.Primary.ID)

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, ac := range r.Accounts {
			if ac.Id == accountID {
				if isTeam {
					rl, err := c.AccountTeams.List(accountID)
					if err != nil {
						if err.(aiven.Error).Status != 404 {
							return err
						}

						return nil
					}

					for _, team := range rl.Teams {
						if team.Id == secondaryID {
							return fmt.Errorf("account team (%s) still exists", rs.Primary.ID)
						}
					}
				} else {
					ra, err := c.AccountAuthentications.List(accountID)
					if err != nil {
						if err.(aiven.Error).Status != 404 {
							return err
						}

						return nil
					}

					for _, a := range ra.AuthenticationMethods {
						if a.Id == secondaryID {
							return fmt.Errorf("account authentication (%s) still exists", rs.Primary.ID)
						}
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
