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

// IsDefaultIPFilter tells if the CIDRs are a list the API sets when ip_filter is not sent, in any order.
func IsDefaultIPFilter(cidrs []string) bool {
	sorted := slices.Sorted(slices.Values(cidrs))
	return slices.ContainsFunc(ipFilterDefaults, func(d []string) bool { return slices.Equal(d, sorted) })
}

// IPFilterPlanModifier plans an IP filter removed from the config as SDKv2 did:
// a default list or null keeps the state value, any other is cleared, which sends "ip_filter": [].
// Skipped on create, when the config sets the field, or another of ipFilterOptions replaces a state value.
func IPFilterPlanModifier() planmodifier.Set {
	return ipFilterPlanModifier{}
}

type ipFilterPlanModifier struct{}

func (ipFilterPlanModifier) Description(context.Context) string {
	return "Clears the IP filter removed from the config, unless the state has a default list or none."
}

func (m ipFilterPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (ipFilterPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, rsp *planmodifier.SetResponse) {
	if !req.ConfigValue.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// The read sets only the options in the plan or the state, e.g., not ip_filter_string for ip_filter
	if req.StateValue.IsNull() {
		rsp.PlanValue = req.StateValue
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
	if len(list) > 0 && !IsDefaultIPFilter(list) {
		rsp.PlanValue = types.SetValueMust(req.StateValue.ElementType(ctx), []attr.Value{})
		return
	}
	rsp.PlanValue = req.StateValue
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
