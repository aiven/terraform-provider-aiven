package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

const defaultPrefix = "test-acc"

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_organization_project", &resource.Sweeper{
		Name: "aiven_organization_project",
		F:    sweepOrganizationProjects(ctx),
	})

	sweep.AddTestSweepers("aiven_organization", &resource.Sweeper{
		Name: "aiven_organization",
		F:    sweepOrganizations(ctx),
	})

	sweep.AddTestSweepers("aiven_organization_application_user", &resource.Sweeper{
		Name: "aiven_organization_application_user",
		F:    sweepOrganizationApplicationUsers(ctx),
		Dependencies: []string{
			"aiven_organization",
		},
	})

	sweep.AddTestSweepers("aiven_organization_user", &resource.Sweeper{
		Name: "aiven_organization_user",
		F:    sweepOrganizationUsers(ctx),
		Dependencies: []string{
			"aiven_organization",
		},
	})

	sweep.AddTestSweepers("aiven_organization_user_group", &resource.Sweeper{
		Name: "aiven_organization_user_group",
		F:    sweepOrganizationUserGroups(ctx),
		Dependencies: []string{
			"aiven_organization",
		},
	})
}

func sweepOrganizations(ctx context.Context) func(string) error {
	return func(_ string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

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

func sweepOrganizationProjects(ctx context.Context) func(string) error {
	return func(_ string) error {
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		organizations, err := client.AccountList(ctx)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of organizations: %w", err)
		}

		if organizations == nil {
			return nil
		}

		for _, organization := range organizations {
			if !strings.HasPrefix(organization.AccountName, "test-acc") {
				continue
			}

			projects, err := client.OrganizationProjectsList(ctx, organization.OrganizationId)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of projects: %w", err)
			}

			for _, project := range projects.Projects {
				if !strings.HasPrefix(project.ProjectId, "test-acc") {
					continue
				}

				if err = client.OrganizationProjectsDelete(ctx, organization.OrganizationId, project.ProjectId); common.IsCritical(err) {
					return fmt.Errorf("error deleting project %s: %w", project.ProjectId, err)
				}
			}
		}

		return nil
	}
}

func sweepOrganizationApplicationUsers(ctx context.Context) func(string) error {
	return func(id string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

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

func sweepOrganizationUserGroups(ctx context.Context) func(string) error {
	return func(id string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

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

func sweepOrganizationUsers(ctx context.Context) func(string) error {
	return func(id string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

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
