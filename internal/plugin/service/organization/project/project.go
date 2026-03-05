package project

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/billinggroup"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func expandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		expandParentID(ctx, client),
		ExpandKeyValue("tag", true),
		billinggroup.ExpandEmails("technical_emails"),
	)
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		flattenParentID,
		FlattenKeyValue("tag"),
		billinggroup.FlattenEmails("technical_emails"),
	)
}

// expandParentID Converts OrganizationID to AccountID in parent_id field because that's what the API expects.
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

// flattenParentID preserves original parent_id if it is an OrganizationID in the state.
// The ParentID in the response is the AccountID,
// while user could have set the OrganizationID in the plan.
// Overrides it with the plan value to avoid an unnecessary diff output.
func flattenParentID(d adapter.ResourceData, dto map[string]any) error {
	if p := d.Get("parent_id").(string); p != "" {
		dto["parent_id"] = p
		return nil
	}
	return nil
}

func ExpandKeyValue(key string, required bool) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		tags, ok := d.GetOk(key)
		if !ok && !required {
			return nil
		}
		m := tags.([]any)
		result := make(map[string]any)
		for _, v := range m {
			kv := v.(map[string]any)
			result[kv["key"].(string)] = kv["value"].(string)
		}
		dto[key] = result
		return nil
	}
}

func FlattenKeyValue(key string) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		tags, ok := dto[key]
		if !ok {
			return nil
		}
		m := tags.(map[string]any)
		list := make([]any, 0)
		for k, v := range m {
			list = append(list, map[string]any{"key": k, "value": v})
		}
		dto[key] = list
		return nil
	}
}
