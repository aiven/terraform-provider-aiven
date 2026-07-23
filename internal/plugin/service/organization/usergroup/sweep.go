package usergroup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_organization_user_group", &resource.Sweeper{
		Name: "aiven_organization_user_group",
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
				if err != nil {
					// AccountList returns all orgs visible to the token, but the token
					// may not have permission on every org.
					// Skip those rather than failing the entire sweep.
					if e, ok := errors.AsType[avngen.Error](err); ok && e.Status == 403 {
						continue
					}

					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving user groups for organization %s: %w", organization.OrganizationId, err)
					}
				}

				for _, group := range groups {
					if !strings.HasPrefix(group.UserGroupName, sweep.DefaultPrefix) {
						continue
					}

					if err = client.UserGroupDelete(ctx, organization.OrganizationId, group.UserGroupId); common.IsCritical(err) {
						return fmt.Errorf("error deleting organization user group %s: %w", group.UserGroupName, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_organization"},
	})
}
