package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_account_team_member", &resource.Sweeper{
		Name:         "aiven_account_team_member",
		F:            sweepAccountTeamMembers(ctx),
		Dependencies: []string{"aiven_account_authentication"},
	})

	sweep.AddTestSweepers("aiven_account_team_project", &resource.Sweeper{
		Name:         "aiven_account_team_project",
		F:            sweepAccountTeamProjects(ctx),
		Dependencies: []string{"aiven_account_authentication"},
	})

	sweep.AddTestSweepers("aiven_account_team", &resource.Sweeper{
		Name:         "aiven_account_team",
		F:            sweepAccountTeams(ctx),
		Dependencies: []string{"aiven_account_team_member", "aiven_account_authentication"},
	})

	sweep.AddTestSweepers("aiven_account", &resource.Sweeper{
		Name:         "aiven_account",
		F:            sweepAccounts(ctx),
		Dependencies: []string{"aiven_project", "aiven_account_team", "aiven_account_team_project", "aiven_account_authentication"},
	})

	sweep.AddTestSweepers("aiven_organizational_unit", &resource.Sweeper{
		Name: "aiven_organizational_unit",
		F:    sweepAccounts(ctx),
	})

	sweep.AddTestSweepers("aiven_account_authentication", &resource.Sweeper{
		Name: "aiven_account_authentication",
		F:    sweepAccountAuthentications(ctx),
	})
}

func listTestAccounts(ctx context.Context) ([]account.AccountOut, error) {
	client, err := sweep.SharedGenClient()
	if err != nil {
		return nil, err
	}

	var testAccounts []account.AccountOut

	resp, err := client.AccountList(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range resp {
		if strings.Contains(a.AccountName, "test-acc-ac-") {
			testAccounts = append(testAccounts, a)
		}
	}

	return testAccounts, nil
}

func sweepAccountAuthentications(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		accounts, err := listTestAccounts(ctx)

		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			aal, err := client.AccountAuthenticationMethodsList(ctx, a.AccountId)
			if err != nil {
				return fmt.Errorf("cannot get account authentications list: %w", err)
			}

			for _, m := range aal {
				if err = client.AccountAuthenticationMethodDelete(ctx, a.AccountId, m.AuthenticationMethodId); err != nil {
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

func sweepAccounts(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		accounts, err := listTestAccounts(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			if err = client.AccountDelete(ctx, a.AccountId); err != nil {
				if common.IsCritical(err) {
					return fmt.Errorf("error destroying account %s during sweep: %w", a.AccountName, err)
				}
			}
		}

		return nil
	}
}

func sweepAccountTeams(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		accounts, err := listTestAccounts(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			atl, err := client.AccountTeamList(ctx, a.AccountId)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %w", err)
			}

			for _, at := range atl {
				if strings.Contains(at.TeamName, "test-acc-team-") {
					err = client.AccountTeamDelete(ctx, a.AccountId, at.TeamId)
					if err != nil {
						return fmt.Errorf("cannot delete account team: %w", err)
					}
				}
			}
		}

		return nil
	}
}

func sweepAccountTeamMembers(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		accounts, err := listTestAccounts(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			atl, err := client.AccountTeamList(ctx, a.AccountId)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %w", err)
			}

			for _, t := range atl {
				if strings.Contains(t.TeamName, "test-acc-team-") {
					// delete all account team invitations
					if t.AccountId == nil {
						return fmt.Errorf("account id is empty for the team %q", t.TeamName)
					}

					mi, err := client.AccountTeamInvitesList(ctx, *t.AccountId, t.TeamId)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team invitations : %w", err)
					}

					for _, i := range mi {
						err = client.AccountTeamMemberCancelInvite(ctx, *t.AccountId, t.TeamId, i.UserEmail)
						if err != nil {
							return fmt.Errorf("cannot delete account team invitation : %w", err)
						}
					}

					// delete all account team members
					tml, err := client.AccountTeamMembersList(ctx, *t.AccountId, t.TeamId)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team members : %w", err)
					}

					for _, tm := range tml {
						if err = client.AccountTeamMembersDelete(ctx, *t.AccountId, t.TeamId, tm.UserId); err != nil {
							return fmt.Errorf("cannot delete account team member : %w", err)
						}
					}
				}
			}
		}

		return nil
	}
}

func sweepAccountTeamProjects(ctx context.Context) func(region string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		accounts, err := listTestAccounts(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of accounts : %w", err)
		}

		for _, a := range accounts {
			atl, err := client.AccountTeamList(ctx, a.AccountId)
			if err != nil {
				return fmt.Errorf("error retrieving a list of account teams : %w", err)
			}

			for _, t := range atl {
				if strings.Contains(t.TeamName, "test-acc-team-") {
					pl, err := client.AccountTeamProjectList(ctx, a.AccountId, t.TeamId)
					if err != nil {
						return fmt.Errorf("error retrieving a list of account team projects : %w", err)
					}

					for _, p := range pl {
						err := client.AccountTeamProjectDisassociate(ctx, a.AccountId, t.TeamId, p.ProjectName)
						if err != nil {
							return fmt.Errorf("cannot delete account team project : %w", err)
						}
					}
				}
			}
		}

		return nil
	}
}
