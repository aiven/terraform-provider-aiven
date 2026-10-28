package adapter

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ipFilterOptions set the same "ip_filter" field of a user config.
var ipFilterOptions = []string{"ip_filter", "ip_filter_string", "ip_filter_object"}

// ipFilterDefaults are the lists the API sets when ip_filter is not sent, sorted.
var ipFilterDefaults = [][]string{
	{"0.0.0.0/0"},
	{"0.0.0.0/0", "::/0"},
}

// IPFilterPlanModifier plans an IP filter removed from the config as SDKv2 did:
// a default list keeps the state value, any other is cleared, which sends "ip_filter": [].
// Skipped when another option sets the field, e.g., ip_filter_string replaces ip_filter.
func IPFilterPlanModifier() planmodifier.Set {
	return ipFilterPlanModifier{}
}

type ipFilterPlanModifier struct{}

func (ipFilterPlanModifier) Description(context.Context) string {
	return "Clears the IP filter removed from the config, unless the state has a default list."
}

func (m ipFilterPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (ipFilterPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, rsp *planmodifier.SetResponse) {
	if !req.ConfigValue.IsNull() || req.StateValue.IsNull() {
		return
	}

	var parent types.Object
	rsp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath(), &parent)...)
	if rsp.Diagnostics.HasError() {
		return
	}

	name := req.Path.Steps()[len(req.Path.Steps())-1].String()
	for k, v := range parent.Attributes() {
		if k != name && slices.Contains(ipFilterOptions, k) && isSetValue(v) {
			return
		}
	}

	list := make([]string, 0, len(req.StateValue.Elements()))
	for _, v := range req.StateValue.Elements() {
		if s, ok := v.(types.String); ok {
			list = append(list, s.ValueString())
		}
	}
	slices.Sort(list)

	if len(list) > 0 && !slices.ContainsFunc(ipFilterDefaults, func(d []string) bool { return slices.Equal(d, list) }) {
		rsp.PlanValue = types.SetValueMust(req.StateValue.ElementType(ctx), []attr.Value{})
	}
}

// isSetValue tells if the config sets a value: unknown, or a collection with items.
func isSetValue(v attr.Value) bool {
	switch {
	case v.IsNull():
		return false
	case v.IsUnknown():
		return true
	}

	switch c := v.(type) {
	case basetypes.SetValue:
		return len(c.Elements()) > 0
	case basetypes.ListValue:
		return len(c.Elements()) > 0
	}
	return true
}
