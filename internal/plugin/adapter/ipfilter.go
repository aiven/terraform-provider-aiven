package adapter

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// AllowIPFilterPurge confirms sending an empty ip_filter, which blocks all the traffic to the service.
const AllowIPFilterPurge = "AIVEN_ALLOW_IP_FILTER_PURGE"

// ipFilterOptions set the same "ip_filter" field of a user config.
var ipFilterOptions = []string{"ip_filter", "ip_filter_string", "ip_filter_object"}

// clearsIPFilter tells if an update sends "ip_filter": [] for an option the config removes, which the state has.
// The plan has no blocks for removed ip_filter_object, as in SDKv2, and IPFilterPlanModifier clears a removed list.
func clearsIPFilter(name string, plan, config, state map[string]any) bool {
	if v, ok := plan[name].(tftypes.Value); ok && !v.IsKnown() {
		return false
	}
	p, _ := asAnySlice(plan[name])
	s, _ := asAnySlice(state[name])
	return slices.Contains(ipFilterOptions, name) && isMapValueUnsetOrEmpty(config, name) && len(p) == 0 && len(s) > 0
}

// checkIPFilterPurge requires AllowIPFilterPurge to send an empty ip_filter.
func checkIPFilterPurge(config map[string]any) error {
	if v, ok := config["ip_filter"].([]any); !ok || len(v) > 0 {
		return nil
	}
	if _, ok := os.LookupEnv(AllowIPFilterPurge); ok {
		return nil
	}
	return fmt.Errorf(
		"ip_filter list is empty, but %[1]s is not set. Please set %[1]s to confirm that you want "+
			"to remove all IP filters, which is going to block all traffic to the service",
		AllowIPFilterPurge,
	)
}

// convertIPFilter fits the items to the option: the API returns them as sent, strings or objects,
// e.g., "10.0.0.0/8" is {"network": "10.0.0.0/8"} in ip_filter_object, and the other way round.
func convertIPFilter(sch *Schema, v any) any {
	list, ok := asAnySlice(v)
	if !ok || sch.Items == nil {
		return v
	}

	result := make([]any, len(list))
	for i, item := range list {
		m, isObject := asAnyMap(item)
		switch {
		case isObject && sch.Items.Type == SchemaTypeString:
			item = m["network"]
		case !isObject && sch.Items.Type == SchemaTypeObject:
			item = map[string]any{"network": item}
		}
		result[i] = item
	}
	return result
}

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
// The state value is kept when the other option has the same CIDRs.
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

	// SDKv2 stored the API value in every option: unknown would update a service the configured option doesn't change
	name := req.Path.Steps()[len(req.Path.Steps())-1].String()
	for k, v := range parent.Attributes() {
		if k != name && slices.Contains(ipFilterOptions, k) && isSetValue(v) {
			if sameNetworks(v, req.StateValue) {
				rsp.PlanValue = req.StateValue
			}
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

// sameNetworks tells if the option has the CIDRs of the state, e.g., ip_filter_object {network = "10.0.0.0/8"} for ["10.0.0.0/8"].
func sameNetworks(option attr.Value, state types.Set) bool {
	if option.IsUnknown() {
		return false
	}

	var items []attr.Value
	switch c := option.(type) {
	case basetypes.SetValue:
		items = c.Elements()
	case basetypes.ListValue:
		items = c.Elements()
	default:
		return false
	}

	got := make([]string, 0, len(items))
	for _, v := range items {
		if o, ok := v.(basetypes.ObjectValue); ok {
			v = o.Attributes()["network"]
		}
		s, ok := v.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			return false
		}
		got = append(got, s.ValueString())
	}

	want := make([]string, 0, len(state.Elements()))
	for _, v := range state.Elements() {
		s, ok := v.(types.String)
		if !ok {
			return false
		}
		want = append(want, s.ValueString())
	}
	slices.Sort(got)
	slices.Sort(want)
	return slices.Equal(slices.Compact(got), slices.Compact(want))
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
