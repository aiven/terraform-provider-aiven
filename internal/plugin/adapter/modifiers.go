package adapter

import (
	"context"
	"fmt"
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

// ExpandUserConfig sends the block of the discriminator value as "user_config", e.g.,
// "datadog_user_config" for "datadog", see userConfigBlock.
// Without the block, sends an empty object, as SDKv2 does: some requests require the field.
func ExpandUserConfig(discriminator string) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		block := userConfigBlock(fmt.Sprint(dto[discriminator]))
		var config map[string]any
		for name := range dto {
			if !isUserConfigBlock(name) {
				continue
			}
			if v, ok := dto[name].(map[string]any); ok && name == block {
				config = v
			}
			delete(dto, name)
		}

		if config == nil {
			config = make(map[string]any)
		}
		dropUnsetObjects(d.Schema().Properties[block], config)
		dto[userConfigField] = config
		return nil
	}
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
