package adapter

import (
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func fromTFValue(sch *Schema, value tftypes.Value) (map[string]any, error) {
	if !value.IsKnown() || value.IsNull() {
		return nil, nil
	}

	val, err := fromTFValueAny(sch, value)
	if err != nil {
		return nil, err
	}

	m, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", val)
	}
	return m, nil
}

func fromTFValueAny(sch *Schema, value tftypes.Value) (any, error) {
	if !value.IsKnown() || value.IsNull() {
		return zeroValue(sch.Type), nil
	}

	switch sch.Type {
	case SchemaTypeString:
		var s string
		err := value.As(&s)
		return s, err
	case SchemaTypeInt:
		var i big.Float
		err := value.As(&i)
		v, _ := i.Int64()
		return int(v), err
	case SchemaTypeFloat:
		var f big.Float
		err := value.As(&f)
		v, _ := f.Float64()
		return v, err
	case SchemaTypeBool:
		var b bool
		err := value.As(&b)
		return b, err
	case SchemaTypeList, SchemaTypeSet:
		var list []tftypes.Value
		if err := value.As(&list); err != nil {
			return nil, err
		}

		if len(list) == 0 {
			return []any{}, nil
		}

		if sch.Items == nil {
			return nil, fmt.Errorf("items is nil")
		}

		result := make([]any, 0, len(list))
		for _, elem := range list {
			item, err := fromTFValueAny(sch.Items, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}

		return result, nil
	case SchemaTypeMap:
		var mapVal map[string]tftypes.Value
		if err := value.As(&mapVal); err != nil {
			return nil, err
		}
		if sch.Items == nil {
			return nil, fmt.Errorf("map items is nil")
		}
		result := make(map[string]any, len(mapVal))
		for key, elem := range mapVal {
			if !elem.IsKnown() || elem.IsNull() {
				continue
			}
			item, err := fromTFValueAny(sch.Items, elem)
			if err != nil {
				return nil, err
			}
			result[key] = item
		}

		if len(result) == 0 {
			return map[string]any{}, nil
		}
		return result, nil
	case SchemaTypeObject:
		var mapVal map[string]tftypes.Value
		if err := value.As(&mapVal); err != nil {
			return nil, err
		}
		result := make(map[string]any, len(mapVal))
		for key, elem := range mapVal {
			if !elem.IsKnown() || elem.IsNull() {
				continue
			}
			itemSch, ok := sch.Properties[key]
			if !ok {
				return nil, fmt.Errorf("unknown property %q", key)
			}

			item, err := fromTFValueAny(itemSch, elem)
			if err != nil {
				return nil, err
			}
			result[key] = item
		}

		if len(result) == 0 {
			return map[string]any{}, nil
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported tftypes.Type: %T", value.Type())
	}
}

func toTFValue(sch *Schema, value any) (tftypes.Value, error) {
	if sch == nil {
		return tftypes.Value{}, fmt.Errorf("schema is nil")
	}

	switch sch.Type {
	case SchemaTypeString:
		if value == nil {
			return tftypes.NewValue(tftypes.String, nil), nil
		}
		v, ok := value.(string)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected string, got %T", value)
		}
		return tftypes.NewValue(tftypes.String, v), nil
	case SchemaTypeInt:
		if value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}
		return tftypes.NewValue(tftypes.Number, value), nil
	case SchemaTypeFloat:
		if value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}
		return tftypes.NewValue(tftypes.Number, value), nil
	case SchemaTypeBool:
		if value == nil {
			return tftypes.NewValue(tftypes.Bool, nil), nil
		}
		v, ok := value.(bool)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected bool, got %T", value)
		}
		return tftypes.NewValue(tftypes.Bool, v), nil
	case SchemaTypeList:
		if sch.Items == nil {
			return tftypes.Value{}, fmt.Errorf("list items is nil")
		}
		nullItem, err := toTFValue(sch.Items, nil)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("failed to build list item type: %w", err)
		}
		listType := tftypes.List{ElementType: nullItem.Type()}
		if value == nil {
			return tftypes.NewValue(listType, nil), nil
		}
		list, ok := value.([]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected []any, got %T", value)
		}

		if len(list) == 0 {
			return tftypes.NewValue(listType, []tftypes.Value{}), nil
		}

		result := make([]tftypes.Value, len(list))
		for i, item := range list {
			converted, err := toTFValue(sch.Items, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("invalid item at index %d: %w", i, err)
			}
			result[i] = converted
		}
		return tftypes.NewValue(listType, result), nil
	case SchemaTypeSet:
		if sch.Items == nil {
			return tftypes.Value{}, fmt.Errorf("set items is nil")
		}
		nullItem, err := toTFValue(sch.Items, nil)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("failed to build set item type: %w", err)
		}
		setType := tftypes.Set{ElementType: nullItem.Type()}
		if value == nil {
			return tftypes.NewValue(setType, nil), nil
		}
		set, ok := value.([]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected []any, got %T", value)
		}

		if len(set) == 0 {
			return tftypes.NewValue(setType, []tftypes.Value{}), nil
		}

		result := make([]tftypes.Value, len(set))
		for i, item := range set {
			converted, err := toTFValue(sch.Items, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("invalid item at index %d: %w", i, err)
			}
			result[i] = converted
		}
		return tftypes.NewValue(setType, result), nil
	case SchemaTypeMap:
		if sch.Items == nil {
			return tftypes.Value{}, fmt.Errorf("map items is nil")
		}
		nullItem, err := toTFValue(sch.Items, nil)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("failed to build map item type: %w", err)
		}
		mapType := tftypes.Map{ElementType: nullItem.Type()}
		if value == nil {
			return tftypes.NewValue(mapType, nil), nil
		}
		m, ok := value.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected map[string]any, got %T", value)
		}

		result := make(map[string]tftypes.Value, len(m))
		for k, item := range m {
			converted, err := toTFValue(sch.Items, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("invalid map value for key %q: %w", k, err)
			}
			result[k] = converted
		}
		return tftypes.NewValue(mapType, result), nil
	case SchemaTypeObject:
		attrs := make(map[string]tftypes.Type, len(sch.Properties))
		for key, prop := range sch.Properties {
			nullProp, err := toTFValue(prop, nil)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("failed to build object property %q type: %w", key, err)
			}
			attrs[key] = nullProp.Type()
		}
		objectType := tftypes.Object{AttributeTypes: attrs}
		if value == nil {
			return tftypes.NewValue(objectType, nil), nil
		}

		m, ok := value.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected map[string]any, got %T", value)
		}

		objectValue := make(map[string]tftypes.Value, len(objectType.AttributeTypes))

		for key, prop := range sch.Properties {
			item, exists := m[key]
			if !exists {
				nullValue, err := toTFValue(prop, nil)
				if err != nil {
					return tftypes.Value{}, fmt.Errorf("invalid object property %q: %w", key, err)
				}
				objectValue[key] = nullValue
				continue
			}

			converted, err := toTFValue(prop, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("invalid object property %q: %w", key, err)
			}
			objectValue[key] = converted
		}

		return tftypes.NewValue(objectType, objectValue), nil
	default:
		return tftypes.Value{}, fmt.Errorf("unsupported schema type: %q", sch.Type)
	}
}
