package project

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func expandModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		expandParentID(ctx, client),
		util.ExpandKeyValueToMap[tfModel](true, "key", "value", "tags"),
		util.ExpandArrayToObjects[tfModel](false, "email", "tech_emails"),
	)
}

func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		flattenParentID,
		util.FlattenMapToKeyValue[tfModel]("key", "value", "tags"),
		util.FlattenObjectsToArray[tfModel]("email", "tech_emails"),
	)
}

// expandParentID Converts OrganizationID to AccountID in parent_id field because that's what the API expects.
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
		return r.Set(parentID, "parent_id")
	}
}

// flattenParentID preserves original parent_id if it is an OrganizationID in the state.
// The ParentID in the response is the AccountID,
// while user could have set the OrganizationID in the plan.
// Overrides it with the plan value to avoid an unnecessary diff output.
func flattenParentID(r util.RawMap, plan *tfModel) error {
	if plan.ParentID.ValueString() != "" {
		return r.Set(plan.ParentID.ValueString(), "parent_id")
	}
	return nil
}
