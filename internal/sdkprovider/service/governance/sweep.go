package governance

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

	sweep.AddTestSweepers("aiven_governance_access", &resource.Sweeper{
		Name: "aiven_governance_access",
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
				acls, err := client.OrganizationGovernanceAccessList(ctx, organization.OrganizationId)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving a list of governance access: %w", err)
				}

				for _, acl := range acls.Access {
					if !strings.HasPrefix(acl.AccessName, defaultPrefix) {
						continue
					}

					if _, err = client.OrganizationGovernanceAccessDelete(ctx, organization.OrganizationId, acl.AccessId); common.IsCritical(err) {
						return fmt.Errorf("error deleting governance access %s: %w", acl.AccessName, err)
					}
				}
			}

			return nil
		},
	})
}
