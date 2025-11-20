package project

import (
	"context"
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func expandModifier(req util.RawMap, state *tfModel) error {
	return util.ComposeModifiers(
		expandParentID,
		// The "tags" field is list of Key-Value map, but in the API it is a map
		util.ExpandKeyValueToMap[tfModel](true, "key", "value", "tags"),
		// Technical emails are defined as array of strings, but in the API it is an array of objects with "email" key.
		util.ExpandArrayToObjects[tfModel](false, "email", "tech_emails"),
	)(req, state)
}

func flattenModifier(r util.RawMap, state *tfModel) error {
	return util.ComposeModifiers(
		flattenParentID,
		util.FlattenMapToKeyValue[tfModel]("key", "value", "tags"),
		util.FlattenObjectsToArray[tfModel]("email", "tech_emails"),
	)(r, state)
}

// expandParentID Converts OrganizationID to AccountID in parent_id field.
func expandParentID(req util.RawMap, plan *tfModel) error {
	pID := plan.ParentID.ValueString()
	if pID == "" {
		return nil
	}

	// This is a workaround. We shouldn't use the client here,
	// but we need to convert the organization ID to an account ID because of legacy code.
	client, err := common.GenClient()
	if err != nil {
		return fmt.Errorf("cannot create aiven client to expand parent_id: %w", err)
	}

	ctx := context.Background()
	parentID, err := schemautil.ConvertOrganizationToAccountID(ctx, pID, client)
	if err != nil {
		return err
	}
	return req.Set(parentID, "parent_id")
}

// flattenParentID preserves original parent_id if it is an OrganizationID.
// The ParentID in the response is the AccountID,
// while user could have set the OrganizationID in the plan.
// Overrides it with the plan value to avoid an unnecessary diff output.
func flattenParentID(r util.RawMap, plan *tfModel) error {
	parentID := plan.ParentID.ValueString()
	if schemautil.IsOrganizationID(parentID) {
		return r.Set(parentID, "parent_id")
	}
	return nil
}
