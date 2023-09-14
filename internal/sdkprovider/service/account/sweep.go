//go:build sweep

package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	client, err := sweep.SharedClient()
	if err != nil {
		panic(fmt.Sprintf("error getting client: %s", err))
	}

	resource.AddTestSweepers("aiven_account_team_member", &resource.Sweeper{
		Name:         "aiven_account_team_member",
		F:            sweepAccountTeamMembers(ctx, client),
		Dependencies: []string{"aiven_account_authentication"},
	})

	resource.AddTestSweepers("aiven_account_team_project", &resource.Sweeper{
		Name:         "aiven_account_team_project",
		F:            sweepAccountTeamProjects(ctx, client),
		Dependencies: []string{"aiven_account_authentication"},
	})

	resource.AddTestSweepers("aiven_account_team", &resource.Sweeper{
		Name:         "aiven_account_team",
		F:            sweepAccountTeams(ctx, client),
		Dependencies: []string{"aiven_account_team_member", "aiven_account_authentication"},
	})

	resource.AddTestSweepers("aiven_account", &resource.Sweeper{
		Name:         "aiven_account",
		F:            sweepAccounts(ctx, client),
		Dependencies: []string{"aiven_project", "aiven_account_team", "aiven_account_team_project", "aiven_account_authentication"},
	})
	resource.AddTestSweepers("aiven_account_authentication", &resource.Sweeper{
		Name: "aiven_account_authentication",
		F:    sweepAccountAuthentications(ctx, client),
	})
}

func listTestAccounts(ctx context.Context, client *aiven.Client) ([]aiven.Account, error) {
	var testAccounts []aiven.Account

	r, err := client.Accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range r.Accounts {
		if strings.Contains(a.Name, "test-acc-ac-") {
			testAccounts = append(testAccounts, a)
		}
	}

	return testAccounts, nil
}

func sweepAccountAuthentications(ctx context.Context, client *aiven.Client) func(region string) error {
	return func(region string) error {
		accounts, err := listTestAccounts(ctx, client)

		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			rr, err := client.AccountAuthentications.List(ctx, a.Id)
			if err != nil {
				return fmt.Errorf("cannot get account authentications list: %w", err)
			}

			for _, m := range rr.AuthenticationMethods {
				err := client.AccountAuthentications.Delete(ctx, a.Id, m.AuthenticationMethodID)
				if err != nil {
					if strings.Contains(err.Error(), "Internal authentication methods cannot be deleted") {
						continue
					}
					return fmt.Errorf("cannot delete account authentication: %w", err)
				}
			}
		}

		return nil
	}
}

func sweepAccounts(ctx context.Context, client *aiven.Client) func(region string) error {
	return func(region string) error {
		accounts, err := listTestAccounts(ctx, client)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			if err := client.Accounts.Delete(ctx, a.Id); err != nil {
				if err.(aiven.Error).Status == 404 {
					continue
				}

				return fmt.Errorf("error destroying account %s during sweep: %w", a.Name, err)
			}
		}

		return nil
	}
}

func sweepAccountTeams(ctx context.Context, client *aiven.Client) func(region string) error {
	return func(region string) error {
		accounts, err := listTestAccounts(ctx, client)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			tr, err := client.AccountTeams.List(ctx, a.Id)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %w", err)
			}

			for _, t := range tr.Teams {
				if strings.Contains(t.Name, "test-acc-team-") {
					err = client.AccountTeams.Delete(ctx, t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("cannot delete account team: %w", err)
					}
				}

			}
		}

		return nil
	}
}
func sweepAccountTeamMembers(ctx context.Context, client *aiven.Client) func(region string) error {
	return func(region string) error {
		accounts, err := listTestAccounts(ctx, client)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %s", err)
		}

		for _, a := range accounts {
			tr, err := client.AccountTeams.List(ctx, a.Id)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %s", err)
			}

			for _, t := range tr.Teams {
				if strings.Contains(t.Name, "test-acc-team-") {
					// delete all account team invitations
					mi, err := client.AccountTeamInvites.List(ctx, t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team invitations : %s", err)
					}

					for _, i := range mi.Invites {
						err := client.AccountTeamInvites.Delete(ctx, i.AccountId, i.TeamId, i.UserEmail)
						if err != nil {
							return fmt.Errorf("cannot delete account team invitation : %s", err)
						}
					}

					// delete all account team members
					mr, err := client.AccountTeamMembers.List(ctx, t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team members : %s", err)
					}

					for _, m := range mr.Members {
						err := client.AccountTeamMembers.Delete(ctx, t.AccountId, t.Id, m.UserId)
						if err != nil {
							return fmt.Errorf("cannot delete account team member : %s", err)
						}
					}
				}

			}
		}

		return nil
	}
}

func sweepAccountTeamProjects(ctx context.Context, client *aiven.Client) func(region string) error {
	return func(region string) error {
		accounts, err := listTestAccounts(ctx, client)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %s", err)
		}

		for _, a := range accounts {
			tr, err := client.AccountTeams.List(ctx, a.Id)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %s", err)
			}

			for _, t := range tr.Teams {
				if strings.Contains(t.Name, "test-acc-team-") {
					pr, err := client.AccountTeamProjects.List(ctx, t.AccountId, t.Id)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team projects : %s", err)
					}

					for _, p := range pr.Projects {
						err := client.AccountTeamProjects.Delete(ctx, t.AccountId, t.Id, p.ProjectName)
						if err != nil {
							return fmt.Errorf("cannot delete account team project : %s", err)
						}
					}
				}

			}
		}

		return nil
	}
}
