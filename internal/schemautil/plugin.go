package schemautil

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liip/sheriff"
)

func ExpandSet[T any](ctx context.Context, diags *diag.Diagnostics, list types.Set) (items []T) {
	if list.IsUnknown() || list.IsNull() {
		return nil
	}
	diags.Append(list.ElementsAs(ctx, &items, false)...)
	return items
}

type Expander[T, K any] func(ctx context.Context, diags *diag.Diagnostics, o *T) *K

func ExpandSetNested[T, K any](ctx context.Context, diags *diag.Diagnostics, expand Expander[T, K], list types.Set) []*K {
	expanded := ExpandSet[T](ctx, diags, list)
	if expanded == nil || diags.HasError() {
		return nil
	}

	items := make([]*K, 0, len(expanded))
	for _, v := range expanded {
		items = append(items, expand(ctx, diags, &v))
		if diags.HasError() {
			return make([]*K, 0)
		}
	}
	return items
}

func ExpandSetBlockNested[T, K any](ctx context.Context, diags *diag.Diagnostics, expand Expander[T, K], list types.Set) *K {
	items := ExpandSetNested(ctx, diags, expand, list)
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

type Flattener[T, K any] func(ctx context.Context, diags *diag.Diagnostics, o *T) *K

func FlattenSetNested[T, K any](ctx context.Context, diags *diag.Diagnostics, flatten Flattener[T, K], attrs map[string]attr.Type, list []*T) types.Set {
	oType := types.ObjectType{AttrTypes: attrs}
	empty := types.SetValueMust(oType, []attr.Value{})
	items := make([]*K, 0, len(list))
	for _, v := range list {
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

func FlattenSetBlockNested[T, K any](ctx context.Context, diags *diag.Diagnostics, flatten Flattener[T, K], attrs map[string]attr.Type, o *T) types.Set {
	if o == nil {
		return types.SetValueMust(types.ObjectType{AttrTypes: attrs}, []attr.Value{})
	}
	return FlattenSetNested(ctx, diags, flatten, attrs, []*T{o})
}

// marshalUserConfig converts user config into json
func marshalUserConfig(c any, groups ...string) (map[string]any, error) {
	if c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()) {
		return nil, nil
	}

	o := &sheriff.Options{
		Groups: groups,
	}

	i, err := sheriff.Marshal(o, c)
	if err != nil {
		return nil, err
	}

	m, ok := i.(map[string]any)
	if !ok {
		// It is an empty pointer
		// sheriff just returned the very same object
		return nil, nil
	}

	return m, nil
}

// MarshalCreateUserConfig returns marshaled user config for Create operation
func MarshalCreateUserConfig(c any) (map[string]any, error) {
	return marshalUserConfig(c, "create", "update")
}

// MarshalUpdateUserConfig returns marshaled user config for Update operation
func MarshalUpdateUserConfig(c any) (map[string]any, error) {
	return marshalUserConfig(c, "update")
}

func MapToDTO(src map[string]any, dst any) error {
	b, err := json.Marshal(&src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// ValueStringPointer checks for "unknown"
// Returns nil instead of zero value
func ValueStringPointer(v types.String) *string {
	if v.IsUnknown() || v.IsNull() {
		return nil
	}
	return v.ValueStringPointer()
}

// ValueBoolPointer checks for "unknown"
// Returns nil instead of zero value
func ValueBoolPointer(v types.Bool) *bool {
	if v.IsUnknown() || v.IsNull() {
		return nil
	}
	return v.ValueBoolPointer()
}

// ValueInt64Pointer checks for "unknown"
// Returns nil instead of zero value
func ValueInt64Pointer(v types.Int64) *int64 {
	if v.IsUnknown() || v.IsNull() {
		return nil
	}
	return v.ValueInt64Pointer()
}

// ValueFloat64Pointer checks for "unknown"
// Returns nil instead of zero value
func ValueFloat64Pointer(v types.Float64) *float64 {
	if v.IsUnknown() || v.IsNull() {
		return nil
	}
	return v.ValueFloat64Pointer()
}
