package unit

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func expandModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		expandParentID(ctx, client),
	)
}

func flattenModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		flattenParentID,
	)
}

// expandParentID if parent_account_id is an OrganizationID, converts it to AccountID for the API request.parent_account_id field.
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
		return r.Set(parentID, "parent_account_id")
	}
}

// flattenParentID preserves parent_account_id format in the state: either OrganizationID or AccountID.
func flattenParentID(r util.RawMap, plan *tfModel) error {
	if plan.ParentID.ValueString() != "" {
		return r.Set(plan.ParentID.ValueString(), "parent_account_id")
	}
	return nil
}

// planModifier user must provide either ID or Name to read the view data source.
// When Name is provided, we need to resolve it to ID first.
// For the datasource only.
func planModifier(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	if state.ID.ValueString() != "" {
		// Resource has ID set, no need to modify the plan.
		return nil
	}

	var diags diag.Diagnostics
	list, err := client.AccountList(ctx)
	if err != nil {
		diags.Append(errmsg.FromError("Plan Modifier Error", err))
		return diags
	}

	for _, a := range list {
		if a.AccountName == state.Name.ValueString() {
			state.SetID(a.AccountId)
			return nil
		}
	}

	diags.AddError("Not Found", "organization unit with the specified name was not found")
	return diags
}
