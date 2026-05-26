package adapter

import (
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func fromTFValue(sch *Schema, value tftypes.Value, preserveNilUnknown bool) (map[string]any, error) {
	if !value.IsKnown() || value.IsNull() {
		return nil, nil
	}

	val, err := fromTFValueAny(sch, value, preserveNilUnknown)
	if err != nil {
		return nil, err
	}

	m, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", val)
	}
	return m, nil
}

func fromTFValueAny(sch *Schema, value tftypes.Value, preserveNilUnknown bool) (any, error) {
	if !value.IsKnown() || value.IsNull() {
		if preserveNilUnknown {
			return value, nil
		}
		return zeroValue(sch.Type), nil
	}

	switch sch.Type {
	case SchemaTypeString:
		var s string
		err := value.As(&s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert string: %w", err)
		}
		return s, nil
	case SchemaTypeInt:
		var i big.Float
		err := value.As(&i)
		if err != nil {
			return nil, fmt.Errorf("failed to convert int: %w", err)
		}
		v, _ := i.Int64()
		return int(v), nil
	case SchemaTypeFloat:
		var f big.Float
		err := value.As(&f)
		if err != nil {
			return nil, fmt.Errorf("failed to convert float: %w", err)
		}
		v, _ := f.Float64()
		return v, nil
	case SchemaTypeBool:
		var b bool
		err := value.As(&b)
		if err != nil {
			return nil, fmt.Errorf("failed to convert bool: %w", err)
		}
		return b, nil
	case SchemaTypeList, SchemaTypeSet:
		var array []tftypes.Value
		if err := value.As(&array); err != nil {
			return nil, fmt.Errorf("failed to convert %s: %w", sch.Type, err)
		}

		if len(array) == 0 {
			return []any{}, nil
		}

		if sch.Items == nil {
			return nil, fmt.Errorf("items is nil")
		}

		result := make([]any, 0, len(array))
		for _, elem := range array {
			item, err := fromTFValueAny(sch.Items, elem, preserveNilUnknown)
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
			if !preserveNilUnknown && (!elem.IsKnown() || elem.IsNull()) {
				continue
			}
			item, err := fromTFValueAny(sch.Items, elem, preserveNilUnknown)
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
			if !preserveNilUnknown && (!elem.IsKnown() || elem.IsNull()) {
				continue
			}
			itemSch, ok := sch.Properties[key]
			if !ok {
				return nil, fmt.Errorf("unknown property %q", key)
			}

			item, err := fromTFValueAny(itemSch, elem, preserveNilUnknown)
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

	// This is a preserved nil/unknown value for ModifyPlan.
	if v, ok := value.(tftypes.Value); ok {
		return v, nil
	}

	if sch.Type.IsPrimitive() {
		t := map[SchemaType]tftypes.Type{
			SchemaTypeString: tftypes.String,
			SchemaTypeInt:    tftypes.Number,
			SchemaTypeFloat:  tftypes.Number,
			SchemaTypeBool:   tftypes.Bool,
		}
		return tftypes.NewValue(t[sch.Type], value), nil
	}

	switch sch.Type {
	case SchemaTypeList, SchemaTypeSet:
		if sch.Items == nil {
			return tftypes.Value{}, fmt.Errorf("%s items is nil", sch.Type)
		}
		nullItem, err := toTFValue(sch.Items, nil)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("failed to build %s item type: %w", sch.Type, err)
		}

		var arrayType tftypes.Type
		if sch.Type == SchemaTypeSet {
			arrayType = tftypes.Set{ElementType: nullItem.Type()}
		} else {
			arrayType = tftypes.List{ElementType: nullItem.Type()}
		}

		if value == nil {
			return tftypes.NewValue(arrayType, nil), nil
		}
		array, ok := asAnySlice(value)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected %s, got %T", sch.Type, value)
		}

		if len(array) == 0 {
			return tftypes.NewValue(arrayType, nil), nil
		}

		result := make([]tftypes.Value, len(array))
		for i, item := range array {
			converted, err := toTFValue(sch.Items, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("invalid item at index %d: %w", i, err)
			}
			result[i] = converted
		}
		return tftypes.NewValue(arrayType, result), nil
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
		m, ok := asAnyMap(value)
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

		m, ok := asAnyMap(value)
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
