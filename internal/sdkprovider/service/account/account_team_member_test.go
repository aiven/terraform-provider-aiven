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
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAccountTeamMember_basic(t *testing.T) {
	t.Skip(accountTeamDeprecated)

	resourceName := "aiven_account_team_member.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountTeamMemberResourceDestroy,
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
	accountName := acc.AccountName()

	return fmt.Sprintf(`
data "aiven_account" "foo" {
  name = "%[1]s"
}

resource "aiven_account_team" "foo" {
  account_id = data.aiven_account.foo.account_id
  name       = "test-acc-team-%[2]s"
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
}`, accountName, name, name)
}

func testAccCheckAivenAccountTeamMemberResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error instantiating client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each account team project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team_member" {
			continue
		}

		accountID, teamID, userEmail, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := c.AccountList(ctx)
		if common.IsCritical(err) {
			return err
		}

		for _, r := range resp {
			if r.AccountId == accountID {
				respTI, err := c.AccountTeamInvitesList(ctx, accountID, teamID)
				if common.IsCritical(err) {
					return err
				}

				for _, invite := range respTI {
					if invite.UserEmail == userEmail {
						return fmt.Errorf("account team member (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}

		for _, r := range resp {
			if r.AccountId == accountID {
				respTM, err := c.AccountTeamMembersList(ctx, accountID, teamID)
				if common.IsCritical(err) {
					return err
				}

				for _, member := range respTM {
					if member.UserEmail == userEmail {
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
