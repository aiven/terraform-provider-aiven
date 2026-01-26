package billinggroup

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.Read = read
	DataSourceOptions.Read = read
}

// read wraps the read operation with ID fallback logic for backward compatibility.
// This handles state migration from SDK provider (<=v4.47).
func read(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// SDK provider stored the ID only in the "id" field, not "billing_group_id"
	// we need to fall back to id for backward compatibility
	billingGroupID := state.BillingGroupID.ValueString()
	if billingGroupID == "" {
		billingGroupID = state.ID.ValueString()
	}
	if billingGroupID == "" {
		diags.AddError("Missing billing group ID", "Neither billing_group_id nor id is set in state")
		return diags
	}

	rsp, err := client.BillingGroupGet(ctx, billingGroupID)
	if err != nil {
		diags.Append(errmsg.FromError("BillingGroupGet Error", err))
		return diags
	}

	diags.Append(flattenData(ctx, state, rsp, flattenModifier(ctx, client))...)

	// convert empty string to null for backward compatibility.
	// SDK provider stored card_id="" in state, but Plugin Framework treats "" and null differently
	// without this normalization, TF would detect a drift
	if state.CardID.ValueString() == "" && !state.CardID.IsNull() {
		state.CardID = types.StringNull()
	}

	return diags
}

func expandModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		getFullCardID(ctx, client),
		expandParentID(ctx, client),
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_contact_emails"),
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_emails"),
	)
}

func flattenModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		flattenParentID(ctx, client),
		util.FlattenObjectsToArray[tfModel]("email", "billing_contact_emails"),
		util.FlattenObjectsToArray[tfModel]("email", "billing_emails"),
	)
}

// getFullCardID turns "last number" card ID into full card ID
func getFullCardID(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		cardID := plan.CardID.ValueString()
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
				return r.Set(card.CardId, "card_id")
			}
		}

		return fmt.Errorf("cannot get card id for %s", cardID)
	}
}

// expandParentID if parent_id is an OrganizationID, converts it to AccountID for the API request.account_id field.
func expandParentID(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		pID := plan.ParentID.ValueString()
		if pID == "" {
			return nil
		}

		parentID, err := schemautil.ConvertOrganizationToAccountID(ctx, pID, client)
		if err != nil {
			return err
		}
		return r.Set(parentID, "account_id")
	}
}

// flattenParentID preserves parent_id format in the state: either OrganizationID or AccountID.
// Copies the logic from util.determineMixedOrganizationConstraintIDToStore
func flattenParentID(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		accountID, _ := r.GetString("account_id")
		if !schemautil.IsOrganizationID(plan.ParentID.ValueString()) {
			// 1. parent_id == "" (not set in the plan)
			// 2. parent_id != orgXXXXXX (is already an account ID)
			return r.Set(accountID, "parent_id")
		}

		a, err := client.AccountGet(ctx, accountID)
		if err != nil {
			return err
		}

		// The field documentation says that parent_id is OrganizationID.
		// But not all accounts used to have OrganizationID set.
		parentID := a.OrganizationId
		if parentID == "" {
			parentID = a.AccountId
		}

		return r.Set(parentID, "parent_id")
	}
}
