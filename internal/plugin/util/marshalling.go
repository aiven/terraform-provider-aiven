package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Unmarshal unmarshals a value from in to out.
// A rebranded copy from SDK package to avoid circular dependency.
func Unmarshal(in, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// CastSlice casts a slice of any to a slice of T.
func CastSlice[T any](v any) ([]T, error) {
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected []any, got %T", v)
	}

	result := make([]T, 0, len(items))
	for _, item := range items {
		v, ok := item.(T)
		if !ok {
			return nil, fmt.Errorf("expected %T element, got %T", v, item)
		}
		result = append(result, v)
	}
	return result, nil
}

// Expand sets values to DTO from a nested attribute/block
// T - terraform object, for instance baseModelAddress
// D - DTO object, for instance a map[string]any
// Expand recursively expands the TF object.
type Expand[T, D any] func(ctx context.Context, diags diag.Diagnostics, obj T) D

// ExpandSet sets values to DTO
func ExpandSet[T any](ctx context.Context, diags diag.Diagnostics, set types.Set) (items []T) {
	if set.IsNull() {
		return nil
	}
	diags.Append(set.ElementsAs(ctx, &items, false)...)
	return items
}

// ExpandSetNestedAttribute sets values to DTO
func ExpandSetNestedAttribute[T, D any](ctx context.Context, diags diag.Diagnostics, expand Expand[T, D], set types.Set) []D {
	// Gets TF objects from the set.
	elements := ExpandSet[T](ctx, diags, set)
	if elements == nil || diags.HasError() {
		return nil
	}

	// Goes deeper and expands each objects
	items := make([]D, 0, len(elements))
	for _, v := range elements {
		// Expands the object using the provided expander function.
		m := expand(ctx, diags, v)
		if diags.HasError() {
			return nil
		}
		items = append(items, m)
	}
	return items
}

// Flatten reads values from DTO for a nested attribute/block
// T - terraform model, for instance baseModelAddress
// D - DTO object, for instance a map[string]any
// Flatten recursively flattens the DTO object.
type Flatten[D, T any] func(ctx context.Context, diags diag.Diagnostics, dto D) T

// FlattenSet reads values from DTO
// Can be used with SetNestedAttribute and SetNestedBlock
func FlattenSet[D, T any](ctx context.Context, diags diag.Diagnostics, flatten Flatten[D, T], set []D, attrs map[string]attr.Type) types.Set {
	oType := types.ObjectType{AttrTypes: attrs}
	empty := types.SetNull(oType)
	if len(set) == 0 {
		return empty
	}

	items := make([]T, 0, len(set))
	for _, v := range set {
		items = append(items, flatten(ctx, diags, v))
		if diags.HasError() {
			return empty
		}
	}

	result, d := types.SetValueFrom(ctx, oType, items)
	diags.Append(d...)
	if diags.HasError() {
		return empty
	}
	return result
}

// FlattenSetNestedAttribute reads values from DTO
// The name reserved for the SetNestedAttribute, as there is also SetNestedBlock which is different
func FlattenSetNestedAttribute[D, T any](ctx context.Context, diags diag.Diagnostics, flatten Flatten[D, T], set []D, attrs map[string]attr.Type) types.Set {
	return FlattenSet(ctx, diags, flatten, set, attrs)
}
