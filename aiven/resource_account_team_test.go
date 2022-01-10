// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_account_team", &resource.Sweeper{
		Name:         "aiven_account_team",
		F:            sweepAccountTeams,
		Dependencies: []string{"aiven_account_team_member"},
	})
}

func sweepAccountTeams(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	r, err := conn.Accounts.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of accounts : %s", err)
	}

	for _, a := range r.Accounts {
		if strings.Contains(a.Name, "test-acc-ac-") {
			tr, err := conn.AccountTeams.List(a.Id)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %s", err)
			}

			for _, t := range tr.Teams {
				if strings.Contains(t.Name, "test-acc-team-") {
					err = conn.AccountTeams.Delete(t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("cannot delete account team: %s", err)
					}
				}

			}
		}
	}

	return nil
}

func TestAccAivenAccountTeam_basic(t *testing.T) {
	resourceName := "aiven_account_team.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountTeamResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountTeamResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountTeamAttributes("data.aiven_account_team.team"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-team-%s", rName)),
				),
			},
		},
	})
}

func testAccAccountTeamResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_account" "foo" {
		  name = "test-acc-ac-%s"
		}
		
		resource "aiven_account_team" "foo" {
		  account_id = aiven_account.foo.account_id
		  name       = "test-acc-team-%s"
		}
		
		data "aiven_account_team" "team" {
		  name       = aiven_account_team.foo.name
		  account_id = aiven_account_team.foo.account_id
		
		  depends_on = [aiven_account_team.foo]
		}`,
		name, name)
}

func testAccCheckAivenAccountTeamResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account team is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team" {
			continue
		}

		accountId, teamId := splitResourceID2(rs.Primary.ID)

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, acc := range r.Accounts {
			if acc.Id == accountId {
				rl, err := c.AccountTeams.List(accountId)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
						return err
					}

					return nil
				}

				for _, team := range rl.Teams {
					if team.Id == teamId {
						return fmt.Errorf("account team (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}

	}

	return nil
}

func testAccCheckAivenAccountTeamAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] account team attributes %v", a)

		if a["account_id"] == "" {
			return fmt.Errorf("expected to get an account id from Aiven")
		}

		if a["team_id"] == "" {
			return fmt.Errorf("expected to get a team_id from Aiven")
		}

		if a["name"] == "" {
			return fmt.Errorf("expected to get a name from Aiven")
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
