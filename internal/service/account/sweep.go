//go:build sweep
// +build sweep

package account

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_account_team_member", &resource.Sweeper{
		Name: "aiven_account_team_member",
		F:    sweepAccountTeamMembers,
	})

	resource.AddTestSweepers("aiven_account_team_project", &resource.Sweeper{
		Name: "aiven_account_team_project",
		F:    sweepAccountTeamProjects,
	})

	resource.AddTestSweepers("aiven_account_team", &resource.Sweeper{
		Name:         "aiven_account_team",
		F:            sweepAccountTeams,
		Dependencies: []string{"aiven_account_team_member"},
	})

	resource.AddTestSweepers("aiven_account", &resource.Sweeper{
		Name:         "aiven_account",
		F:            sweepAccounts,
		Dependencies: []string{"aiven_project", "aiven_account_team", "aiven_account_team_project"},
	})
}

func sweepAccounts(region string) error {
	client, err := sweep.SharedClient(region)
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
			if err := conn.Accounts.Delete(a.Id); err != nil {
				if err.(aiven.Error).Status == 404 {
					continue
				}

				return fmt.Errorf("error destroying account %s during sweep: %s", a.Name, err)
			}
		}
	}

	return nil
}

func sweepAccountTeams(region string) error {
	client, err := sweep.SharedClient(region)
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
func sweepAccountTeamMembers(region string) error {
	client, err := sweep.SharedClient(region)
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
					// delete all account team invitations
					mi, err := conn.AccountTeamInvites.List(t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team invitations : %s", err)
					}

					for _, i := range mi.Invites {
						err := conn.AccountTeamInvites.Delete(i.AccountId, i.TeamId, i.UserEmail)
						if err != nil {
							return fmt.Errorf("cannot delete account team invitation : %s", err)
						}
					}

					// delete all account team members
					mr, err := conn.AccountTeamMembers.List(t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team members : %s", err)
					}

					for _, m := range mr.Members {
						err := conn.AccountTeamMembers.Delete(t.AccountId, t.Id, m.UserId)
						if err != nil {
							return fmt.Errorf("cannot delete account team member : %s", err)
						}
					}
				}

			}
		}
	}

	return nil
}

func sweepAccountTeamProjects(region string) error {
	client, err := sweep.SharedClient(region)
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
