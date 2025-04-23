package util

import (
	"context"
	"encoding/json"
	"fmt"

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

// Expand sets values to Request from a nested attribute/block
// T - terraform object, for instance baseModelAddress
// R - Request, for instance, a map[string]any
// Expand recursively expands the TF object.
type Expand[T, R any] func(ctx context.Context, obj T) (R, diag.Diagnostics)

// ExpandSet sets values to Request
func ExpandSet[T any](ctx context.Context, set types.Set) ([]T, diag.Diagnostics) {
	if set.IsNull() {
		return nil, nil
	}

	var items []T
	diags := set.ElementsAs(ctx, &items, false)
	if diags.HasError() {
		return nil, diags
	}
	return items, nil
}

// ExpandSetNested sets values to Request
func ExpandSetNested[T, R any](ctx context.Context, expand Expand[T, R], set types.Set) ([]R, diag.Diagnostics) {
	// Gets TF objects from the set.
	elements, diags := ExpandSet[T](ctx, set)
	if elements == nil || diags.HasError() {
		return nil, diags
	}

	// Goes deeper and expands each object
	items := make([]R, 0, len(elements))
	for _, v := range elements {
		// Expands the object using the provided expander function.
		m, diags := expand(ctx, v)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, m)
	}
	return items, nil
}

// Flatten reads values from Response for a nested attribute/block
// T - terraform model, for instance baseModelAddress
// R - Response, for instance a map[string]any
// Flatten recursively flattens Response object.
type Flatten[R, T any] func(ctx context.Context, response R) (T, diag.Diagnostics)

// FlattenSet reads values from Response
// Can be used with SetNestedAttribute and SetNestedBlock
func FlattenSet[R, T any](ctx context.Context, flatten Flatten[R, T], set []R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	empty := types.SetNull(oType)
	if len(set) == 0 {
		return empty, nil
	}

	items := make([]T, 0, len(set))
	for _, v := range set {
		item, diags := flatten(ctx, v)
		if diags.HasError() {
			return empty, diags
		}
		items = append(items, item)
	}

	result, diags := types.SetValueFrom(ctx, oType, items)
	if diags.HasError() {
		return empty, diags
	}
	return result, nil
}

// FlattenSetNested reads values from Response
func FlattenSetNested[R, T any](ctx context.Context, flatten Flatten[R, T], set []R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	return FlattenSet(ctx, flatten, set, oType)
}

// MapModifier modifies Request and Response objects
type MapModifier func(map[string]any) error
