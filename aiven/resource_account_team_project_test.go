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
	resource.AddTestSweepers("aiven_account_team_project", &resource.Sweeper{
		Name: "aiven_account_team_project",
		F:    sweepAccountTeamProjects,
	})
}

func sweepAccountTeamProjects(region string) error {
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
					pr, err := conn.AccountTeamProjects.List(t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team projects : %s", err)
					}

					for _, p := range pr.Projects {
						err := conn.AccountTeamProjects.Delete(t.AccountId, t.Id, p.ProjectName)
						if err != nil {
							return fmt.Errorf("cannot delete account team project : %s", err)
						}
					}
				}

			}
		}
	}

	return nil
}

func TestAccAivenAccountTeamProject_basic(t *testing.T) {
	resourceName := "aiven_account_team_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
			name = "test-acc-team-%s"
		}

		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			account_id = aiven_account_team.foo.account_id
		}

		resource "aiven_account_team_project" "foo" {
			account_id = aiven_account.foo.account_id
			team_id = aiven_account_team.foo.team_id
			project_name = aiven_project.foo.project
			team_type = "admin"
		}

		data "aiven_account_team_project" "project" {
  			team_id = aiven_account_team_project.foo.team_id
  			account_id = aiven_account_team_project.foo.account_id
			project_name = aiven_account_team_project.foo.project_name

			depends_on = [aiven_account_team_project.foo]
		}
		`, name, name, name)
}

func testAccCheckAivenAccountTeamProjectResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each account team project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team_project" {
			continue
		}

		accountId, teamId, projectName := splitResourceID3(rs.Primary.ID)

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, acc := range r.Accounts {
			if acc.Id == accountId {
				rp, err := c.AccountTeamProjects.List(accountId, teamId)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
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
