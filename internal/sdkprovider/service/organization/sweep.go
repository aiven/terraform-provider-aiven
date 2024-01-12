//go:build sweep

package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

const defaultPrefix = "test-acc"

func init() {
	ctx := context.Background()

	client, err := sweep.SharedClient()
	if err != nil {
		panic(fmt.Sprintf("error getting client: %s", err))
	}

	resource.AddTestSweepers("aiven_organization", &resource.Sweeper{
		Name: "aiven_organization",
		F:    sweepOrganizations(ctx, client),
	})

	resource.AddTestSweepers("aiven_organization_application_user", &resource.Sweeper{
		Name: "aiven_organization_application_user",
		F:    sweepOrganizationApplicationUsers(ctx, client),
		Dependencies: []string{
			"aiven_organization",
		},
	})

	resource.AddTestSweepers("aiven_organization_user", &resource.Sweeper{
		Name: "aiven_organization_user",
		F:    sweepOrganizationUsers(ctx, client),
		Dependencies: []string{
			"aiven_organization",
		},
	})

	resource.AddTestSweepers("aiven_organization_user_group", &resource.Sweeper{
		Name: "aiven_organization_user_group",
		F:    sweepOrganizationUserGroups(ctx, client),
		Dependencies: []string{
			"aiven_organization",
		},
	})

}

func sweepOrganizations(ctx context.Context, client *aiven.Client) func(string) error {
	return func(id string) error {
		organizations, err := client.Accounts.List(ctx)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of organizations: %w", err)
		}

		if organizations == nil {
			return nil
		}

		for _, organization := range organizations.Accounts {
			if !strings.HasPrefix(organization.Name, "test-acc") {
				continue
			}

			err = client.Accounts.Delete(ctx, organization.Id)
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting organization %s: %w", organization.Name, err)
			}
		}

		return nil
	}
}

func sweepOrganizationApplicationUsers(ctx context.Context, client *aiven.Client) func(string) error {
	return func(id string) error {
		organizationApplicationUsers, err := client.OrganizationApplicationUserHandler.List(ctx, id)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of organization application users: %w", err)
		}

		if organizationApplicationUsers == nil {
			return nil
		}

		for _, organizationApplicationUser := range organizationApplicationUsers.Users {
			if !strings.HasPrefix(organizationApplicationUser.Name, defaultPrefix) {
				continue
			}

			err = client.OrganizationApplicationUserHandler.Delete(ctx, id, organizationApplicationUser.UserID)
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting organization application user %s: %w", organizationApplicationUser.Name, err)
			}
		}

		return nil
	}
}

func sweepOrganizationUserGroups(ctx context.Context, client *aiven.Client) func(string) error {
	return func(id string) error {
		organizationUserGroups, err := client.OrganizationUserGroups.List(ctx, id)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of organization user groups: %w", err)
		}

		if organizationUserGroups == nil {
			return nil
		}

		for _, organizationUserGroup := range organizationUserGroups.UserGroups {
			if !strings.HasPrefix(organizationUserGroup.UserGroupName, defaultPrefix) {
				continue
			}

			err = client.OrganizationUserGroups.Delete(ctx, id, organizationUserGroup.UserGroupID)
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting organization user group %s: %w", organizationUserGroup.UserGroupName, err)
			}
		}

		return nil
	}
}

func sweepOrganizationUsers(ctx context.Context, client *aiven.Client) func(string) error {
	return func(id string) error {
		organizationUsers, err := client.OrganizationUser.List(ctx, id)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of organization users: %w", err)
		}

		if organizationUsers == nil {
			return nil
		}

		for _, organizationUser := range organizationUsers.Users {
			if !strings.Contains(organizationUser.UserInfo.UserEmail, defaultPrefix) {
				continue
			}

			err = client.OrganizationUser.Delete(ctx, id, organizationUser.UserID)
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting organization user %s: %w", organizationUser.UserID, err)
			}
		}

		return nil
	}
}
