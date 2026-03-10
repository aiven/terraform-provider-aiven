package applicationusertoken

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_organization_application_user_token", &resource.Sweeper{
		Name: "aiven_organization_application_user_token",
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
				users, err := client.ApplicationUsersList(ctx, organization.OrganizationId)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving application users for organization %s: %w", organization.OrganizationId, err)
				}

				for _, user := range users {
					tokens, err := client.ApplicationUserAccessTokensList(ctx, organization.OrganizationId, user.UserId)
					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving tokens for application user %s: %w", user.UserId, err)
					}

					for _, token := range tokens {
						if token.Description == nil || !strings.HasPrefix(*token.Description, sweep.DefaultPrefix) {
							continue
						}

						if err = client.ApplicationUserAccessTokenDelete(ctx, organization.OrganizationId, user.UserId, token.TokenPrefix); common.IsCritical(err) {
							return fmt.Errorf("error deleting token %s for user %s: %w", token.TokenPrefix, user.UserId, err)
						}
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_organization_application_user"},
	})
}
