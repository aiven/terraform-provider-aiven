package account_test

import (
	"errors"
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

func TestAccAivenAccountTeamProject_basic(t *testing.T) {
	resourceName := "aiven_account_team_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAccountTeamProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountTeamProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountTeamProejctAttributes("data.aiven_account_team_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project_name", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "team_type", "admin"),
				),
			},
		},
	})
}

func testAccAccountTeamProjectResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name       = "test-acc-team-%s"
}

resource "aiven_project" "foo" {
  project    = "test-acc-pr-%s"
  account_id = aiven_account_team.foo.account_id
}

resource "aiven_account_team_project" "foo" {
  account_id   = aiven_account.foo.account_id
  team_id      = aiven_account_team.foo.team_id
  project_name = aiven_project.foo.project
  team_type    = "admin"
}

data "aiven_account_team_project" "project" {
  team_id      = aiven_account_team_project.foo.team_id
  account_id   = aiven_account_team_project.foo.account_id
  project_name = aiven_account_team_project.foo.project_name

  depends_on = [aiven_account_team_project.foo]
}`, name, name, name)
}

func testAccCheckAivenAccountTeamProjectResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account team project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team_project" {
			continue
		}

		accountID, teamID, projectName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
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
			if a.Id == accountID {
				rp, err := c.AccountTeamProjects.List(accountID, teamID)
				if err != nil {
					var aivenError *aiven.Error

					if ok := errors.As(err, &aivenError); !ok || aivenError.Status != 404 {
						return err
					}

					return nil
				}

				for _, p := range rp.Projects {
					if p.ProjectName == projectName {
						return fmt.Errorf("account team project (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountTeamProejctAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] account team project attributes %v", a)

		if a["account_id"] == "" {
			return fmt.Errorf("expected to get an account id from Aiven")
		}

		if a["team_id"] == "" {
			return fmt.Errorf("expected to get a team_id from Aiven")
		}

		if a["project_name"] == "" {
			return fmt.Errorf("expected to get a project_name from Aiven")
		}

		if a["team_type"] == "" {
			return fmt.Errorf("expected to get a team_type from Aiven")
		}

		return nil
	}
}
