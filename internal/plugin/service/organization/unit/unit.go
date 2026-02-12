package unit

import (
	"context"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func expandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		expandParentID(ctx, client),
	)
}

func flattenModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		flattenParentID,
	)
}

// expandParentID if parent_account_id is an OrganizationID, converts it to AccountID for the API request.parent_account_id field.
func expandParentID(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		pID := d.Get("parent_id").(string)
		if pID == "" {
			return nil
		}

		parentID, err := schemautil.ConvertOrganizationToAccountID(ctx, pID, client)
		if err != nil {
			return err
		}
		dto["parent_id"] = parentID
		return nil
	}
}

// flattenParentID preserves parent_account_id format in the state: either OrganizationID or AccountID.
func flattenParentID(d adapter.ResourceData, dto map[string]any) error {
	if p := d.Get("parent_id").(string); p != "" {
		dto["parent_id"] = p
		return nil
	}
	return nil
}

// planModifier user must provide either ID or Name to read the view data source.
// When Name is provided, we need to resolve it to ID first.
// For the datasource only.
func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.ID() != "" {
		// Resource has ID set, no need to modify the plan.
		return nil
	}

	list, err := client.AccountList(ctx)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	for _, a := range list {
		if a.AccountName == name {
			return d.SetID(a.AccountId)
		}
	}

	return &avngen.Error{Status: http.StatusNotFound, Message: "organization unit with the specified name was not found"}
}
