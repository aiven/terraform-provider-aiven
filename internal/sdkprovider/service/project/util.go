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
