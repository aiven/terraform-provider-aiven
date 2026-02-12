package billinggroup

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	// SDK provider stored the ID only in the "id" field, not "billing_group_id"
	// we need to fall back to id for backward compatibility
	if d.Get("billing_group_id") == "" {
		if d.ID() == "" {
			return fmt.Errorf("missing billing group ID: neither billing_group_id nor id is set in state")
		}

		err := d.Set("billing_group_id", d.ID())
		if err != nil {
			return err
		}
	}
	return nil
}

func expandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		getFullCardID(ctx, client),
		expandParentID(ctx, client),
		ExpandEmails("billing_contact_emails"),
		ExpandEmails("billing_emails"),
	)
}

func ExpandEmails(path string) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		v, ok := d.GetOk(path)
		if !ok {
			return nil
		}

		list := v.([]any)
		for i, email := range list {
			list[i] = map[string]any{"email": email.(string)}
		}

		dto[path] = list
		return nil
	}
}

func flattenModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		flattenCardID,
		flattenParentID(ctx, client),
		FlattenEmails("billing_contact_emails"),
		FlattenEmails("billing_emails"),
	)
}

func FlattenEmails(path string) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		v, ok := dto[path]
		if !ok {
			return nil
		}

		list := v.([]any)
		for i, item := range list {
			list[i] = item.(map[string]any)["email"].(string)
		}

		dto[path] = list
		return nil
	}
}

// getFullCardID turns "last number" card ID into full card ID
func getFullCardID(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		cardID := d.Get("card_id").(string)
		if cardID == "" {
			return nil
		}

		//nolint:staticcheck // linter ignore: intended for backwards compatibility
		list, err := client.UserCreditCardsList(ctx)
		if err != nil {
			return fmt.Errorf("cannot get credit cards list: %w", err)
		}

		for _, card := range list {
			if card.CardId == cardID || card.Last4 == cardID {
				dto["card_id"] = card.CardId
				return nil
			}
		}

		return fmt.Errorf("cannot get card id for %s", cardID)
	}
}

// expandParentID if parent_id is an OrganizationID, converts it to AccountID for the API request.account_id field.
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
		dto["account_id"] = parentID
		return nil
	}
}

// flattenCardID
// Converts empty string to null for backward compatibility.
// SDK provider stored card_id="" in state, but Plugin Framework treats "" and null differently
// without this normalization, TF would detect a drift
func flattenCardID(d adapter.ResourceData, dto map[string]any) error {
	if d.Get("card_id").(string) == "" {
		return d.Set("card_id", nil)
	}
	return nil
}

// flattenParentID preserves parent_id format in the state: either OrganizationID or AccountID.
// Copies the logic from util.determineMixedOrganizationConstraintIDToStore
func flattenParentID(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		parentID := d.Get("parent_id").(string)
		if parentID == "" {
			return nil
		}

		if schemautil.IsOrganizationID(parentID) {
			accountID, ok := dto["account_id"].(string)
			if !ok {
				return nil
			}

			a, err := client.AccountGet(ctx, accountID)
			if err != nil {
				return err
			}

			if a.OrganizationId != "" {
				dto["parent_id"] = a.OrganizationId
			}
		}

		return nil
	}
}
