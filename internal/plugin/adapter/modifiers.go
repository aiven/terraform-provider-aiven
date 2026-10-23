package adapter

import "fmt"

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

// ExpandUserConfig sends the block of the discriminator value as the field, e.g.,
// "datadog_user_config" as "user_config" for "datadog". blocks maps discriminator values to block names.
// Without the block, a required field is sent empty, and an optional one is omitted.
func ExpandUserConfig(discriminator, field string, blocks map[string]string, required bool) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		kind := fmt.Sprint(dto[discriminator])
		var config map[string]any
		for k, name := range blocks {
			if v, ok := dto[name].(map[string]any); ok && k == kind {
				config = v
			}
			delete(dto, name)
		}

		switch {
		case config != nil:
			dto[field] = config
		case required:
			dto[field] = make(map[string]any)
		}
		return nil
	}
}

// FlattenUserConfig is the reverse of ExpandUserConfig.
// The API returns the defaults too, so a resource keeps only the options set in the plan or the state,
// and no block when none is set. Otherwise, the defaults show up as diffs, or as blocks missing in the config.
// A data source keeps all the options, and sets no block for an empty field.
func FlattenUserConfig(discriminator, field string, blocks map[string]string) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		config, _ := dto[field].(map[string]any)
		delete(dto, field)

		kind, ok := dto[discriminator]
		if !ok {
			kind = d.Get(discriminator)
		}

		name, ok := blocks[fmt.Sprint(kind)]
		if !ok {
			return nil
		}

		if d.IsResource() {
			if v, ok := keepSetValues(d.Schema().Properties[name], config, d.Get(name)); ok {
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
// UseStateForUnknown plans the options of a new list object as null, not unknown,
// so the defaults the API returns for them would be an inconsistent result.
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
