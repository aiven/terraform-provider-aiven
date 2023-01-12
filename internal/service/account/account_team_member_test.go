package account_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAccountTeamMember_basic(t *testing.T) {
	resourceName := "aiven_account_team_member.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountTeamMemberResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountTeamMemberResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountTeamMemberAttributes("data.aiven_account_team_member.member"),
					resource.TestCheckResourceAttr(resourceName, "user_email", fmt.Sprintf("ivan.savciuc+%s@aiven.fi", rName)),
					resource.TestCheckResourceAttr(resourceName, "accepted", "false"),
				),
			},
		},
	})
}

func testAccAccountTeamMemberResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name       = "test-acc-team-%s"
}

resource "aiven_account_team_member" "foo" {
  team_id    = aiven_account_team.foo.team_id
  account_id = aiven_account_team.foo.account_id
  user_email = "ivan.savciuc+%s@aiven.fi"
}

data "aiven_account_team_member" "member" {
  team_id    = aiven_account_team_member.foo.team_id
  account_id = aiven_account_team_member.foo.account_id
  user_email = aiven_account_team_member.foo.user_email

  depends_on = [aiven_account_team_member.foo]
}`, name, name, name)
}

func testAccCheckAivenAccountTeamMemberResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account team project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team_member" {
			continue
		}

		accountID, teamID, userEmail, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, a := range r.Accounts {
			if a.Id == accountID {
				ri, err := c.AccountTeamInvites.List(accountID, teamID)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
						return err
					}

					return nil
				}

				for _, i := range ri.Invites {
					if i.UserEmail == userEmail {
						return fmt.Errorf("account team member (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountTeamMemberAttributes(n string) resource.TestCheckFunc {
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

		if a["user_email"] == "" {
			return fmt.Errorf("expected to get a user email from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["accepted"] != "false" {
			return fmt.Errorf("expected to get a accepted from Aiven")
		}

		if a["invited_by_user_email"] == "" {
			return fmt.Errorf("expected to get a invited_by_user_email from Aiven")
		}

		return nil
	}
}
