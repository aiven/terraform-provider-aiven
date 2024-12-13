package project

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// accountIDPointer returns the account ID pointer to use in a request to the Aiven API.
// This is limited to the domain of the billing groups and projects.
// If the parent_id is set, it will be used as the account ID. Otherwise, the account_id will be used.
func accountIDPointer(ctx context.Context, client *aiven.Client, d *schema.ResourceData) (*string, error) {
	var accountID *string

	ownerEntityID, ok := d.GetOk("parent_id")
	if ok {
		ownerEntityID, err := schemautil.NormalizeOrganizationID(ctx, client, ownerEntityID.(string))
		if err != nil {
			return nil, err
		}

		if len(ownerEntityID) == 0 {
			return nil, nil
		}

		accountID = &ownerEntityID
	} else {
		// TODO: Remove this when account_id is removed.
		accountID = schemautil.OptionalStringPointer(d, "account_id")
	}

	return accountID, nil
}

// determineMixedOrganizationConstraintIDToStore is a helper function that returns the ID to store in the state.
// We have several fields that can be either an organization ID or an account ID.
// We want to store the one that was already in the state, if it was already there.
// If it was not, we want to prioritize the organization ID, but if it is not available, we want to store the account
// ID.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
// If the ID is an organization ID, it will be refreshed via the provided account ID and returned.
func determineMixedOrganizationConstraintIDToStore(
	ctx context.Context,
	client *aiven.Client,
	stateID string,
	accountID string,
) (string, error) {
	if accountID == "" {
		return "", nil
	}

	if !schemautil.IsOrganizationID(stateID) {
		return accountID, nil
	}

	r, err := client.Accounts.Get(ctx, accountID)
	if err != nil {
		return "", err
	}

	if r.Account.OrganizationId != "" {
		return r.Account.OrganizationId, nil
	}

	return accountID, nil
}
