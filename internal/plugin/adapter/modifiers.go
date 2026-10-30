package adapter

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
)

type MapModifier func(d ResourceData, dto map[string]any) error

func ComposeMapModifiers(modifiers ...MapModifier) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		for _, modifier := range modifiers {
			if err := modifier(d, dto); err != nil {
				return err
			}
		}
		return nil
	}
}

// RenameFields renames field names terraform name -> dto name
func RenameFields(fields map[string]string) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		for was, want := range fields {
			v, ok := dto[was]
			if ok {
				dto[want] = v
				delete(dto, was)
			}
		}
		return nil
	}
}

// SetFixedValues sets request fields from key-value pairs, overriding the plan.
// Used for fields the user can't set, e.g. the discriminator of a polymorphic endpoint.
func SetFixedValues(kv ...string) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		if len(kv)%2 != 0 {
			return fmt.Errorf("SetFixedValues expects key-value pairs, got %d arguments", len(kv))
		}
		for i := 0; i < len(kv); i += 2 {
			dto[kv[i]] = kv[i+1]
		}
		return nil
	}
}

// AllowIPFilterPurge confirms sending an empty ip_filter, which blocks all the traffic to the service.
const AllowIPFilterPurge = "AIVEN_ALLOW_IP_FILTER_PURGE"

// ExpandUserConfig sends the block of the discriminator value as "user_config", e.g.,
// "datadog_user_config" for "datadog", see userConfigBlock.
// Without the block, sends an empty object, as SDKv2 does: some requests require the field.
// The options get their API names, see expandOptions. An update sends the configured options, see keepConfigured.
func ExpandUserConfig(discriminator string) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		block := userConfigBlock(fmt.Sprint(dto[discriminator]))
		sch := d.Schema().Properties[block]
		var config map[string]any
		for name := range dto {
			if !isUserConfigBlock(name) {
				continue
			}
			if v, ok := dto[name].(map[string]any); ok && name == block {
				config = v
				if !d.IsNewResource() && sch != nil {
					c, _ := d.GetConfigOk(name)
					keepConfigured(sch.Items, config, firstObject(c), firstObject(d.GetState(name)))
				}
			}
			delete(dto, name)
		}

		if config == nil {
			config = make(map[string]any)
		}
		dropUnsetObjects(sch, config)
		if sch != nil {
			config = expandOptions(sch.Items, config, d.IsNewResource())
		}

		if v, ok := config["ip_filter"].([]any); ok && len(v) == 0 {
			if _, ok := os.LookupEnv(AllowIPFilterPurge); !ok {
				return fmt.Errorf(
					"ip_filter list is empty, but %[1]s is not set. Please set %[1]s to confirm that you want "+
						"to remove all IP filters, which is going to block all traffic to the service",
					AllowIPFilterPurge,
				)
			}
		}
		dto[userConfigField] = config
		return nil
	}
}

// keepConfigured drops the options removed from the config, which the plan has from the state, as SDKv2 sent the config only.
// The API keeps them, e.g., removing alerting_enabled = true sends nothing, and the API keeps true.
// Except the lists cleared, unless another option sets the same field: removing all the ip_filter_object blocks
// sends "ip_filter": [], and so does an ip_filter that IPFilterPlanModifier clears.
func keepConfigured(sch *Schema, plan, config, state map[string]any) {
	for k, p := range sch.Properties {
		_, configured := config[k]
		stateList, _ := asAnySlice(state[k])
		switch {
		case p.IsObject:
			if c, ok := plan[k].(map[string]any); ok && configured {
				keepConfigured(p.Items, c, firstObject(config[k]), firstObject(state[k]))
			} else {
				delete(plan, k)
			}
		case isFieldConfigured(sch, config, k):
			delete(plan, k)
		case p.Items != nil && p.Items.Type == SchemaTypeObject && !p.IsAttribute:
			if len(stateList) > 0 && isMapValueUnsetOrEmpty(config, k) {
				plan[k] = []any{}
			} else if !configured {
				delete(plan, k)
			}
		case !configured:
			if planList, ok := asAnySlice(plan[k]); !ok || len(planList) > 0 || len(stateList) == 0 {
				delete(plan, k)
			}
		}
	}
}

// isFieldConfigured tells if another option of the API field has items, e.g., ip_filter_string for ip_filter.
func isFieldConfigured(sch *Schema, config map[string]any, name string) bool {
	field := cmp.Or(sch.Properties[name].JSONName, name)
	for k, p := range sch.Properties {
		if k != name && cmp.Or(p.JSONName, k) == field && !isMapValueUnsetOrEmpty(config, k) {
			return true
		}
	}
	return false
}

// expandOptions renames the options to the API names, e.g., {"foo__dot__bar": 1} is sent as {"foo.bar": 1}.
// Drops the create-only options on update: the API accepts them on create only, e.g., admin_password.
func expandOptions(sch *Schema, config map[string]any, isNew bool) map[string]any {
	result := make(map[string]any, len(config))
	for k, v := range config {
		p, ok := sch.Properties[k]
		switch {
		case !ok:
			result[k] = v
		case p.CreateOnly && !isNew:
		default:
			v = mapObjects(p, v, func(s *Schema, m map[string]any) any {
				return expandOptions(s, m, isNew)
			})

			// An empty block, e.g., public_access {}, isn't sent, as SDKv2 did
			if m, ok := v.(map[string]any); ok && p.IsObject && len(m) == 0 {
				continue
			}
			result[cmp.Or(p.JSONName, k)] = v
		}
	}
	return result
}

// flattenOptions renames the API options to the Terraform names, e.g., {"foo.bar": 1} is {"foo__dot__bar": 1}.
// The options of the same API name get the value each, e.g., ip_filter is ip_filter_string too,
// and a resource keeps the one set, see keepSetValues.
func flattenOptions(sch *Schema, config map[string]any) map[string]any {
	result := make(map[string]any, len(config))
	for k, p := range sch.Properties {
		name := cmp.Or(p.JSONName, k)
		v, ok := config[name]
		if !ok {
			continue
		}
		if name == "ip_filter" {
			v = convertIPFilter(p, v)
		}
		result[k] = mapObjects(p, v, func(s *Schema, m map[string]any) any { return flattenOptions(s, m) })
	}
	return result
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

// mapObjects applies f to an object, or to the objects of a list, e.g., to each ip_filter_object item.
func mapObjects(sch *Schema, v any, f func(*Schema, map[string]any) any) any {
	if sch.IsObject {
		sch = sch.Items
	}

	if m, ok := asAnyMap(v); ok && sch.Type == SchemaTypeObject {
		return f(sch, m)
	}

	list, ok := asAnySlice(v)
	if !ok || sch.Items == nil || sch.Items.Type != SchemaTypeObject {
		return v
	}

	result := make([]any, len(list))
	for i, item := range list {
		result[i] = mapObjects(sch.Items, item, f)
	}
	return result
}

// firstObject returns the object of a block, which the state keeps as a list: [{"a": 1}] is {"a": 1}.
func firstObject(v any) map[string]any {
	if list, ok := asAnySlice(v); ok {
		if len(list) == 0 {
			return nil
		}
		v = list[0]
	}
	m, _ := asAnyMap(v)
	return m
}

// dropUnsetObjects removes the nested objects missing in the config: the plan has them as empty lists,
// which the API expects as objects.
func dropUnsetObjects(sch *Schema, value any) {
	if sch == nil {
		return
	}
	if sch.IsObject {
		// Expand unwraps the objects
		sch = sch.Items
	}

	switch v := value.(type) {
	case map[string]any:
		for k, item := range v {
			prop := sch.Properties[k]
			if list, ok := item.([]any); ok && len(list) == 0 && prop != nil && prop.IsObject {
				delete(v, k)
				continue
			}
			dropUnsetObjects(prop, item)
		}
	case []any:
		for _, item := range v {
			dropUnsetObjects(sch.Items, item)
		}
	}
}

// ValidateUserConfig fails when the config sets a block of another user config type,
// e.g. "datadog_user_config" when the discriminator value is "logs".
// The API never gets it, and Terraform rejects the state without it as an inconsistent result after apply.
func ValidateUserConfig(discriminator string) func(context.Context, avngen.Client, ResourceData) error {
	return func(_ context.Context, _ avngen.Client, d ResourceData) error {
		kind, ok := d.GetOk(discriminator)
		if !ok {
			// Unknown until apply
			return nil
		}

		block := userConfigBlock(fmt.Sprint(kind))
		for name := range d.Schema().Properties {
			if name == block || !isUserConfigBlock(name) {
				continue
			}
			if list, _ := asAnySlice(d.Get(name)); len(list) > 0 {
				return fmt.Errorf("%s can't be set when %s is %q", name, discriminator, kind)
			}
		}
		return nil
	}
}

const (
	userConfigField  = "user_config"  // The API field of every type, e.g., {"endpoint_type": "datadog", "user_config": {...}}
	userConfigSuffix = "_user_config" // The suffix of a block name, e.g., "datadog_user_config"
)

// userConfigBlock names the block of a user config type "<type>_user_config", or "<type>" for the legacy blocks.
func userConfigBlock(kind string) string {
	switch kind {
	// The endpoint types whose blocks SDKv2 named without the suffix.
	// Kept for compatibility, never add more: the definition renames the blocks to match.
	// todo: rename to "<type>_user_config" in v5.0.0.
	case "external_google_cloud_bigquery", "external_postgresql":
		return kind
	}
	return kind + userConfigSuffix
}

// isUserConfigBlock reports whether the field is a user config block. A legacy block is named as its type.
func isUserConfigBlock(name string) bool {
	return strings.HasSuffix(name, userConfigSuffix) || userConfigBlock(name) == name
}

// FlattenUserConfig is the reverse of ExpandUserConfig.
// The API returns the defaults too, so a resource keeps only the options set in the plan or the state,
// and no block when none is set. Otherwise, the defaults show up as diffs, or as blocks missing in the config.
// A data source keeps all the options, and sets no block for an empty field.
func FlattenUserConfig(discriminator string) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		config, _ := dto[userConfigField].(map[string]any)
		delete(dto, userConfigField)

		kind, ok := dto[discriminator]
		if !ok {
			kind = d.Get(discriminator)
		}

		name := userConfigBlock(fmt.Sprint(kind))
		sch, ok := d.Schema().Properties[name]
		if !ok {
			return nil
		}

		config = flattenOptions(sch.Items, config)
		if d.IsResource() {
			if v, ok := keepSetValues(sch, config, d.Get(name)); ok {
				dto[name] = v
			}
			return nil
		}

		if len(config) > 0 {
			dto[name] = config
		}
		return nil
	}
}

// keepSetValues returns the response value with only the object fields set in known,
// the plan or the state value of the same schema. The objects of a list are matched by index.
// Other lists, sets and maps are kept whole. Returns false when known is not set.
// The create-only options keep the known value: the API might not return them, e.g., admin_password.
func keepSetValues(sch *Schema, rsp, known any) (any, bool) {
	if sch == nil {
		return nil, false
	}

	if sch.IsObject {
		list, _ := asAnySlice(known)
		if len(list) == 0 {
			return nil, false
		}
		known, sch = list[0], sch.Items
	}

	if known == nil {
		return nil, false
	}

	switch {
	case sch.Type == SchemaTypeList && sch.Items != nil && sch.Items.Type == SchemaTypeObject:
		return keepSetItems(sch.Items, rsp, known), true
	case sch.Type != SchemaTypeObject:
		return rsp, true
	}

	knownObj, _ := asAnyMap(known)
	rspObj, _ := asAnyMap(rsp)
	result := make(map[string]any, len(knownObj))
	for k, v := range rspObj {
		if v == nil {
			continue
		}
		if v, ok := keepSetValues(sch.Properties[k], v, knownObj[k]); ok {
			result[k] = v
		}
	}

	for k, v := range knownObj {
		if p := sch.Properties[k]; p != nil && p.CreateOnly && v != nil && result[k] == nil {
			result[k] = v
		}
	}
	return result, true
}

// keepSetItems filters each response object by the known object of the same index.
// The response objects without a known one are dropped, e.g. the default clickhouse_postgresql databases.
func keepSetItems(sch *Schema, rsp, known any) []any {
	rspList, _ := asAnySlice(rsp)
	knownList, _ := asAnySlice(known)
	result := make([]any, min(len(rspList), len(knownList)))
	for i := range result {
		result[i] = rspList[i]
		if kept, ok := keepSetValues(sch, rspList[i], knownList[i]); ok {
			result[i] = kept
		}
	}
	return result
}
