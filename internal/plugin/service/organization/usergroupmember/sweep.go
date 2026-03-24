package usergroupmember

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/go-client-codegen/handler/usergroup"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_organization_user_group_member", &resource.Sweeper{
		Name: "aiven_organization_user_group_member",
		F: func(_ string) error {
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
				groups, err := client.UserGroupsList(ctx, organization.OrganizationId)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving user groups for organization %s: %w", organization.OrganizationId, err)
				}

				for _, group := range groups {
					if !strings.HasPrefix(group.UserGroupName, sweep.DefaultPrefix) {
						continue
					}

					members, err := client.UserGroupMemberList(ctx, organization.OrganizationId, group.UserGroupId)
					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving members for group %s: %w", group.UserGroupName, err)
					}

					for _, member := range members {
						err = client.UserGroupMembersUpdate(ctx, organization.OrganizationId, group.UserGroupId, &usergroup.UserGroupMembersUpdateIn{
							MemberIds: []string{member.UserId},
							Operation: usergroup.OperationTypeRemoveMembers,
						})
						if common.IsCritical(err) {
							return fmt.Errorf("error removing member %s from group %s: %w", member.UserId, group.UserGroupName, err)
						}
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_organization_user_group"},
	})
}
