package organizationbillinggroup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_organization_billing_group", &resource.Sweeper{
		Name: "aiven_organization_billing_group",
		F: func(_ string) error {
			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}

			organizations, err := client.AccountList(ctx)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of organizations: %w", err)
			}

			for _, organization := range organizations {
				billingGroups, err := client.OrganizationBillingGroupList(ctx, organization.OrganizationId)
				if err != nil {
					// The token may not have permission on every org returned by AccountList.
					if e, ok := errors.AsType[avngen.Error](err); ok && e.Status == http.StatusForbidden {
						continue
					}

					if common.IsCritical(err) {
						return fmt.Errorf("error retrieving billing groups for organization %s: %w", organization.OrganizationId, err)
					}
				}

				for _, bg := range billingGroups {
					// The backend creates "Default billing group for <project>" for test projects
					// without a billing group and keeps it after the project is deleted.
					name := strings.TrimPrefix(bg.BillingGroupName, "Default billing group for ")
					if !strings.HasPrefix(name, sweep.DefaultPrefix) {
						continue
					}

					err = client.OrganizationBillingGroupDelete(ctx, organization.OrganizationId, bg.BillingGroupId)
					if common.IsCritical(err) {
						return fmt.Errorf("error deleting billing group %s (%s): %w", bg.BillingGroupName, bg.BillingGroupId, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{
			"aiven_project",
		},
	})
}
