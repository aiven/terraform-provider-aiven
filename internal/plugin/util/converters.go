package util

import (
	"bytes"
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

	// Uses json.Number to avoid int->float64->int overflow issue.
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	return d.Decode(out)
}

// Remarshal remarshals a value from in to out, applies typed modifiers before unmarshalling to out.
func Remarshal[I any, O any](in *I, out *O, modifiers ...MapModifier[I]) error {
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
	var diags diag.Diagnostics
	items := make([]*R, 0, len(target))
	for _, v := range target {
		// Expands the object using the provided expander function.
		m, d := expand(ctx, &v)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, m)
	}
	return items, diags
}

// ExpandSetNested recursively sets values to Request for a set of objects.
func ExpandSetNested[T, R any](ctx context.Context, expand Expand[T, R], set types.Set) ([]*R, diag.Diagnostics) {
	var target []T
	diags := set.ElementsAs(ctx, &target, false)
	if diags.HasError() {
		return nil, diags
	}

	items, d := expandEach(ctx, expand, target)
	diags.Append(d...)
	if len(items) == 0 || diags.HasError() {
		return nil, diags
	}
	return items, diags
}

// ExpandSingleNested sets values to Request for a single object.
func ExpandSingleNested[T, R any](ctx context.Context, expand Expand[T, R], list types.List) (*R, diag.Diagnostics) {
	var target []T
	diags := list.ElementsAs(ctx, &target, false)
	if diags.HasError() {
		return nil, diags
	}

	items, d := expandEach(ctx, expand, target)
	diags.Append(d...)
	if len(items) == 0 || diags.HasError() {
		return nil, diags
	}
	return items[0], diags
}

func ExpandMapNested[T, R any](ctx context.Context, expand Expand[T, R], dict types.Map) (map[string]*R, diag.Diagnostics) {
	target := make(map[string]T)
	diags := dict.ElementsAs(ctx, &target, false)
	if len(target) == 0 || diags.HasError() {
		return nil, diags
	}

	elements := make(map[string]*R, len(target))
	for k, v := range target {
		// Expands the object using the provided expander function.
		m, d := expand(ctx, &v)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		elements[k] = m
	}
	return elements, diags
}

// Flatten reads values from Response for a nested object.
// T - terraform model, for instance baseModelAddress
// R - Response DTO, for instance foo.GetOut.
// Flatten recursively flattens Response object.
type Flatten[R, T any] func(ctx context.Context, response *R) (*T, diag.Diagnostics)

// FlattenSetNested reads values from Response for a set of objects.
// Same as types.SetValueFrom() returns unknown type on error.
func FlattenSetNested[R, T any](ctx context.Context, flatten Flatten[R, T], set []*R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	if len(set) == 0 {
		return types.SetNull(oType), nil
	}

	var diags diag.Diagnostics
	unknown := types.SetUnknown(oType)

	items := make([]*T, 0, len(set))
	for _, v := range set {
		item, d := flatten(ctx, v)
		diags.Append(d...)
		if diags.HasError() {
			return unknown, diags
		}
		items = append(items, item)
	}

	result, d := types.SetValueFrom(ctx, oType, items)
	diags.Append(d...)
	if diags.HasError() {
		return unknown, diags
	}
	return result, diags
}

// FlattenSingleNested reads values from Response for a single object.
// Same as types.ListValueFrom() returns unknown type on error.
func FlattenSingleNested[R, T any](ctx context.Context, flatten Flatten[R, T], dto *R, oType types.ObjectType) (types.List, diag.Diagnostics) {
	if dto == nil {
		return types.ListNull(oType), nil
	}

	var diags diag.Diagnostics
	unknown := types.ListUnknown(oType)

	item, d := flatten(ctx, dto)
	diags.Append(d...)
	if diags.HasError() {
		return unknown, diags
	}

	result, d := types.ListValueFrom(ctx, oType, []*T{item})
	diags.Append(d...)
	if diags.HasError() {
		return unknown, diags
	}
	return result, diags
}

// FlattenMapNested reads values from Response for a map of objects.
// Same as types.MapValueFrom() returns unknown type on error.
func FlattenMapNested[R, T any](ctx context.Context, flatten Flatten[R, T], dict map[string]*R, oType types.ObjectType) (types.Map, diag.Diagnostics) {
	if len(dict) == 0 {
		return types.MapNull(oType), nil
	}

	var diags diag.Diagnostics
	unknown := types.MapUnknown(oType)

	elements := make(map[string]*T, len(dict))
	for k, v := range dict {
		item, d := flatten(ctx, v)
		diags.Append(d...)
		if diags.HasError() {
			return unknown, diags
		}
		elements[k] = item
	}

	result, d := types.MapValueFrom(ctx, oType, elements)
	diags.Append(d...)
	if diags.HasError() {
		return unknown, diags
	}
	return result, diags
}
