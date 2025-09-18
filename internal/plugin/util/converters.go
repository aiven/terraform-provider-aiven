package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// remarshal unmarshals a value from in to out.
// A rebranded copy from SDK package to avoid circular dependency.
func remarshal(in, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// Unmarshal unmarshals a value from in to out, applies typed modifiers before unmarshalling to out.
func Unmarshal[I any, O any](in *I, out *O, modifiers ...MapModifier[I]) error {
	if len(modifiers) == 0 {
		return remarshal(in, out)
	}

	b, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	m := NewRawMap(b)
	for _, modify := range modifiers {
		err := modify(m, in)
		if err != nil {
			return fmt.Errorf("failed to apply modifier: %w", err)
		}
	}

	// Leave this branch for debugging purposes.
	err = json.Unmarshal(m.Data(), out)
	if err != nil {
		return err
	}
	return nil
}

// MapModifier modifies Request and Response objects
// reqRsp — a map that would be used to unmarshal a Request/Response object
// in *I — input type, for instance, a dataModel or foo.CreateIn dto.
// Hint: to ignore "in" type and modify only the map, use `any`, for instance:
//
//	func foo[T any](reqRsp RawMap, _ *T) error {
//		reqRsp["foo"] = "bar"
//	}
//
// In this case, the function can be used with any type of Request/Response object.
// The map's keys are equal to original API keys, not TF schema fields.
type MapModifier[I any] func(reqRsp RawMap, in *I) error

// Expand sets values to Request from a nested object.
// T - terraform object, for instance baseModelAddress
// R - Request DTO, for instance, foo.CreateIn dto.
// Expand recursively expands the TF object (dataModel).
type Expand[T, R any] func(ctx context.Context, obj *T) (*R, diag.Diagnostics)

func expandEach[T, R any](ctx context.Context, expand Expand[T, R], target []T) ([]*R, diag.Diagnostics) {
	// Goes deeper and expands each object
	items := make([]*R, 0, len(target))
	for _, v := range target {
		// Expands the object using the provided expander function.
		m, diags := expand(ctx, &v)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, m)
	}
	return items, nil
}

// ExpandSetNested recursively sets values to Request for a set of objects.
func ExpandSetNested[T, R any](ctx context.Context, expand Expand[T, R], set types.Set) ([]*R, diag.Diagnostics) {
	var target []T
	diags := set.ElementsAs(ctx, &target, false)
	if diags.HasError() {
		return nil, diags
	}
	return expandEach(ctx, expand, target)
}

// ExpandSingleNested sets values to Request for a single object.
func ExpandSingleNested[T, R any](ctx context.Context, expand Expand[T, R], list types.List) (*R, diag.Diagnostics) {
	var target []T
	diags := list.ElementsAs(ctx, &target, false)
	if diags.HasError() || len(target) == 0 {
		return nil, diags
	}

	items, diags := expandEach(ctx, expand, target)
	if diags.HasError() || len(items) == 0 {
		return nil, diags
	}

	return items[0], nil
}

func ExpandMapNested[T, R any](ctx context.Context, expand Expand[T, R], dict types.Map) (map[string]*R, diag.Diagnostics) {
	target := make(map[string]T)
	diags := dict.ElementsAs(ctx, &target, false)
	if diags.HasError() {
		return nil, diags
	}

	elements := make(map[string]*R, len(target))
	for k, v := range target {
		// Expands the object using the provided expander function.
		m, diags := expand(ctx, &v)
		if diags.HasError() {
			return nil, diags
		}
		elements[k] = m
	}
	return elements, nil
}

// Flatten reads values from Response for a nested object.
// T - terraform model, for instance baseModelAddress
// R - Response DTO, for instance foo.GetOut.
// Flatten recursively flattens Response object.
type Flatten[R, T any] func(ctx context.Context, response *R) (*T, diag.Diagnostics)

// FlattenSetNested reads values from Response for a set of objects.
func FlattenSetNested[R, T any](ctx context.Context, flatten Flatten[R, T], set []*R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	null := types.SetNull(oType)
	if len(set) == 0 {
		return null, nil
	}

	items := make([]*T, 0, len(set))
	for _, v := range set {
		item, diags := flatten(ctx, v)
		if diags.HasError() {
			return null, diags
		}
		items = append(items, item)
	}

	result, diags := types.SetValueFrom(ctx, oType, items)
	if diags.HasError() {
		return null, diags
	}
	return result, nil
}

// FlattenSingleNested reads values from Response for a single object.
func FlattenSingleNested[R, T any](ctx context.Context, flatten Flatten[R, T], dto *R, oType types.ObjectType) (types.List, diag.Diagnostics) {
	null := types.ListNull(oType)
	if dto == nil {
		return null, nil
	}

	item, diags := flatten(ctx, dto)
	if diags.HasError() {
		return null, diags
	}

	result, diags := types.ListValueFrom(ctx, oType, []*T{item})
	if diags.HasError() {
		return null, diags
	}
	return result, nil
}

func FlattenMapNested[R, T any](ctx context.Context, flatten Flatten[R, T], dict map[string]*R, oType types.ObjectType) (types.Map, diag.Diagnostics) {
	null := types.MapNull(oType)
	if len(dict) == 0 {
		return null, nil
	}

	elements := make(map[string]*T, len(dict))
	for k, v := range dict {
		item, diags := flatten(ctx, v)
		if diags.HasError() {
			return null, diags
		}
		elements[k] = item
	}

	result, diags := types.MapValueFrom(ctx, oType, elements)
	if diags.HasError() {
		return null, diags
	}
	return result, nil
}
