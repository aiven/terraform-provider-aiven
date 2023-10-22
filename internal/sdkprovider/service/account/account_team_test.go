package account_test

import (
	"context"
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

func TestAccAivenAccountTeam_basic(t *testing.T) {
	if _, ok := os.LookupEnv("AIVEN_ACCOUNT_NAME"); !ok {
		t.Skip("AIVEN_ACCOUNT_NAME env variable is required to run this test")
	}

	resourceName := "aiven_account_team.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountTeamResourceDestroy,
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
	orgName := os.Getenv("AIVEN_ACCOUNT_NAME")

	return fmt.Sprintf(`
data "aiven_account" "foo" {
  name = "%s"
}

resource "aiven_account_team" "foo" {
  account_id = data.aiven_account.foo.account_id
  name       = "test-acc-team-%s"
}

data "aiven_account_team" "team" {
  name       = aiven_account_team.foo.name
  account_id = aiven_account_team.foo.account_id

  depends_on = [aiven_account_team.foo]
}`, orgName, name)
}

func testAccCheckAivenAccountTeamResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each account team is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team" {
			continue
		}

		accountID, teamID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.Accounts.List(ctx)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, ac := range r.Accounts {
			if ac.Id == accountID {
				rl, err := c.AccountTeams.List(ctx, accountID)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
						return err
					}

					return nil
				}

				for _, team := range rl.Teams {
					if team.Id == teamID {
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
