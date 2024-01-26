package account_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAccountTeamMember_basic(t *testing.T) {
	if _, ok := os.LookupEnv("AIVEN_ACCOUNT_NAME"); !ok {
		t.Skip("AIVEN_ACCOUNT_NAME env variable is required to run this test")
	}

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
	orgName := os.Getenv("AIVEN_ACCOUNT_NAME")

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
  user_email = "ivan.savciuc+%[2]s@aiven.fi"
}

data "aiven_account_team_member" "member" {
  team_id    = aiven_account_team_member.foo.team_id
  account_id = aiven_account_team_member.foo.account_id
  user_email = aiven_account_team_member.foo.user_email

  depends_on = [aiven_account_team_member.foo]
}`, orgName, name)
}

func testAccCheckAivenAccountTeamMemberResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

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

		r, err := c.Accounts.List(ctx)
		if err != nil {
			var e *aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}

			return nil
		}

		for _, a := range r.Accounts {
			if a.Id == accountID {
				ri, err := c.AccountTeamInvites.List(ctx, accountID, teamID)
				if err != nil {
					var e *aiven.Error
					if errors.As(err, &e) && e.Status != 404 {
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
