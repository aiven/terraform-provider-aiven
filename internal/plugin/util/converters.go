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

	// To applied modifiers, we need to unmarshal into a map[string]any
	m := make(map[string]any)
	err := remarshal(in, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal map: %w", err)
	}

	for _, modify := range modifiers {
		err := modify(m, in)
		if err != nil {
			return fmt.Errorf("failed to apply modifier: %w", err)
		}
	}

	return remarshal(&m, out)
}

// MapModifier modifies Request and Response objects
// reqRsp — a map that would be used to unmarshal a Request/Response object
// in *I — input type, for instance, a dataModel or foo.CreateIn dto.
// Hint: to ignore "in" type and modify only the map, use `any`, for instance:
//
//	func foo[T any](reqRsp map[string]any, _ *T) error {
//		reqRsp["foo"] = "bar"
//	}
//
// In this case, the function can be used with any type of Request/Response object.
type MapModifier[I any] func(reqRsp map[string]any, in *I) error

// Expand sets values to Request from a nested object.
// T - terraform object, for instance baseModelAddress
// R - Request DTO, for instance, foo.CreateIn dto.
// Expand recursively expands the TF object (dataModel).
type Expand[T, R any] func(ctx context.Context, obj *T) (*R, diag.Diagnostics)

// ExpandSet sets values to Request for a set of objects/scalars.
func ExpandSet[T any](ctx context.Context, set types.Set) ([]T, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}

	var items []T
	diags := set.ElementsAs(ctx, &items, false)
	if diags.HasError() {
		return nil, diags
	}
	return items, nil
}

// ExpandSetNested recursively sets values to Request for a set of objects.
func ExpandSetNested[T, R any](ctx context.Context, expand Expand[T, R], set types.Set) ([]*R, diag.Diagnostics) {
	// Gets TF objects from the set.
	elements, diags := ExpandSet[T](ctx, set)
	if elements == nil || diags.HasError() {
		return nil, diags
	}

	// Goes deeper and expands each object
	items := make([]*R, 0, len(elements))
	for _, v := range elements {
		// Expands the object using the provided expander function.
		m, diags := expand(ctx, &v)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, m)
	}
	return items, nil
}

// ExpandSingle sets values to Request for a single object.
func ExpandSingle[T, R any](ctx context.Context, expand Expand[T, R], set types.Set) (*R, diag.Diagnostics) {
	result, diags := ExpandSetNested(ctx, expand, set)
	if diags.HasError() || len(result) == 0 {
		return nil, diags
	}
	return result[0], nil
}

// Flatten reads values from Response for a nested object.
// T - terraform model, for instance baseModelAddress
// R - Response DTO, for instance foo.GetOut.
// Flatten recursively flattens Response object.
type Flatten[R, T any] func(ctx context.Context, response *R) (*T, diag.Diagnostics)

// FlattenSetNested reads values from Response for a set of objects.
func FlattenSetNested[R, T any](ctx context.Context, flatten Flatten[R, T], set []*R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	empty := types.SetNull(oType)
	if len(set) == 0 {
		return empty, nil
	}

	items := make([]*T, 0, len(set))
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

// FlattenSingle reads values from Response for a single object.
func FlattenSingle[R, T any](ctx context.Context, flatten Flatten[R, T], dto *R, oType types.ObjectType) (types.Set, diag.Diagnostics) {
	if dto == nil {
		return types.SetNull(oType), nil
	}
	return FlattenSetNested(ctx, flatten, []*R{dto}, oType)
}
