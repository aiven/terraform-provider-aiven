package acctest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// defaultBillingGroupPrefix names the billing group the backend creates for a project
// added to an organization without one. Deleting the project does not delete it.
const defaultBillingGroupPrefix = "Default billing group for "

// TestAccCheckBillingGroupsDestroyed fails if an aiven_organization_billing_group from the state
// still exists, or if the backend kept the default billing group of an aiven_project from the state.
func TestAccCheckBillingGroupsDestroyed(s *terraform.State) error {
	c, err := GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aiven_organization_billing_group":
			orgID, billingGroupID, err := schemautil.SplitResourceID2(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = c.OrganizationBillingGroupGet(ctx, orgID, billingGroupID)
			if isGone(err) {
				continue
			}
			if err != nil {
				return fmt.Errorf("error getting organization billing group %s: %w", rs.Primary.ID, err)
			}
			return fmt.Errorf("organization billing group (%s) still exists", rs.Primary.ID)
		case "aiven_project":
			err = checkDefaultBillingGroupDestroyed(ctx, c, rs.Primary)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkDefaultBillingGroupDestroyed(ctx context.Context, c avngen.Client, project *terraform.InstanceState) error {
	parentID := project.Attributes["parent_id"]
	if parentID == "" {
		parentID = project.Attributes["account_id"]
	}
	if parentID == "" {
		return nil
	}

	orgID := parentID
	if !schemautil.IsOrganizationID(parentID) {
		account, err := c.AccountGet(ctx, parentID)
		if isGone(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting account %s: %w", parentID, err)
		}
		orgID = account.OrganizationId
	}

	billingGroups, err := c.OrganizationBillingGroupList(ctx, orgID)
	if isGone(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing billing groups of organization %s: %w", orgID, err)
	}

	name := defaultBillingGroupPrefix + project.ID
	for _, bg := range billingGroups {
		if bg.BillingGroupName == name {
			return fmt.Errorf("billing group %q (%s) of project %s still exists", name, bg.BillingGroupId, project.ID)
		}
	}
	return nil
}

// isGone reports whether err means the resource or its organization no longer exists.
func isGone(err error) bool {
	if avngen.IsNotFound(err) {
		return true
	}
	e, ok := errors.AsType[avngen.Error](err)
	return ok && e.Status == http.StatusForbidden
}
