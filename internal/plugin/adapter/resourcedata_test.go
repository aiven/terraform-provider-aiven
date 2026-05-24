// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestDereference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		value           any
		want            any
		wantErrorSubstr string
	}{
		{name: "nil", value: nil, want: nil},
		{name: "string", value: "hello", want: "hello"},
		{name: "int", value: 42, want: 42},
		{name: "bool", value: true, want: true},
		{name: "float64", value: 3.14, want: 3.14},
		{name: "slice", value: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "map", value: map[string]int{"x": 1}, want: map[string]int{"x": 1}},
		{name: "pointer to string", value: new("hello"), want: "hello"},
		{name: "pointer to int", value: new(42), want: 42},
		{name: "pointer to bool", value: new(true), want: true},
		{name: "pointer to float64", value: new(3.14), want: 3.14},
		{name: "pointer to slice", value: new([]string{"a", "b"}), want: []string{"a", "b"}},
		{name: "pointer to map", value: new(map[string]int{"x": 1}), want: map[string]int{"x": 1}},
		{name: "pointer to pointer to string", value: new(new("hello")), wantErrorSubstr: "pointer to pointer"},
		{name: "3 pointers to string", value: new(new(new("hello"))), wantErrorSubstr: "pointer to pointer"},
		{name: "nil pointer", value: (*string)(nil), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dereference(tt.value)
			if tt.wantErrorSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrorSubstr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestTFValue_NormalizeEmpties verifies the two symmetric rules applied by
// tfValue() to keep plan and config in agreement on empty containers:
//
//  1. Lift: config has [] or {}, plan is missing/nil → encoded as empty.
//  2. Drop: plan has [] or {}, config is missing/nil → encoded as null.
//
// Without these rules, Terraform rejects the apply with "produced
// inconsistent result after apply".
func TestTFValue_NormalizeEmpties(t *testing.T) {
	t.Parallel()

	strSch := &Schema{Type: SchemaTypeString}
	nestedSch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"name": strSch,
			"tags": {Type: SchemaTypeList, Items: strSch},
		},
	}
	rootSch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":     strSch,
			"tags":   {Type: SchemaTypeList, Items: strSch},
			"set":    {Type: SchemaTypeSet, Items: strSch},
			"labels": {Type: SchemaTypeMap, Items: strSch},
			"nested": nestedSch,
		},
	}

	encode := func(t *testing.T, plan, config map[string]any) tftypes.Value {
		t.Helper()
		rd, err := NewResourceDataFromMaps(rootSch, []string{"id"}, plan, nil, config)
		require.NoError(t, err)
		return rd.(*resourceData).tfValue()
	}

	attr := func(t *testing.T, v tftypes.Value, name string) tftypes.Value {
		t.Helper()
		var m map[string]tftypes.Value
		require.NoError(t, v.As(&m))
		got, ok := m[name]
		require.True(t, ok, "attribute %q missing", name)
		return got
	}

	t.Run("lift: empty list in config, missing in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		config := map[string]any{"id": "x", "tags": []any{}}

		got := attr(t, encode(t, plan, config), "tags")
		require.False(t, got.IsNull())
		require.True(t, got.IsKnown())

		var list []tftypes.Value
		require.NoError(t, got.As(&list))
		require.Empty(t, list)
	})

	t.Run("lift: empty set in config, missing in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		config := map[string]any{"id": "x", "set": []any{}}

		got := attr(t, encode(t, plan, config), "set")
		require.False(t, got.IsNull())

		var s []tftypes.Value
		require.NoError(t, got.As(&s))
		require.Empty(t, s)
	})

	t.Run("lift: empty map in config, missing in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		config := map[string]any{"id": "x", "labels": map[string]any{}}

		got := attr(t, encode(t, plan, config), "labels")
		require.False(t, got.IsNull())

		var m map[string]tftypes.Value
		require.NoError(t, got.As(&m))
		require.Empty(t, m)
	})

	t.Run("lift: empty list in config, nil in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x", "tags": nil}
		config := map[string]any{"id": "x", "tags": []any{}}

		got := attr(t, encode(t, plan, config), "tags")
		require.False(t, got.IsNull(), "nil in plan must be treated like a missing key")

		var list []tftypes.Value
		require.NoError(t, got.As(&list))
		require.Empty(t, list)
	})

	t.Run("lift: empty set in config, nil in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x", "set": nil}
		config := map[string]any{"id": "x", "set": []any{}}

		got := attr(t, encode(t, plan, config), "set")
		require.False(t, got.IsNull())

		var s []tftypes.Value
		require.NoError(t, got.As(&s))
		require.Empty(t, s)
	})

	t.Run("lift: empty map in config, nil in plan, encodes as empty", func(t *testing.T) {
		plan := map[string]any{"id": "x", "labels": nil}
		config := map[string]any{"id": "x", "labels": map[string]any{}}

		got := attr(t, encode(t, plan, config), "labels")
		require.False(t, got.IsNull())

		var m map[string]tftypes.Value
		require.NoError(t, got.As(&m))
		require.Empty(t, m)
	})

	t.Run("drop: empty list in plan, missing in config, encodes as null", func(t *testing.T) {
		plan := map[string]any{"id": "x", "tags": []any{}}
		config := map[string]any{"id": "x"}

		got := attr(t, encode(t, plan, config), "tags")
		require.True(t, got.IsNull(), "empty plan list must be dropped when config does not declare it")
	})

	t.Run("drop: empty map in plan, missing in config, encodes as null", func(t *testing.T) {
		plan := map[string]any{"id": "x", "labels": map[string]any{}}
		config := map[string]any{"id": "x"}

		got := attr(t, encode(t, plan, config), "labels")
		require.True(t, got.IsNull(), "empty plan map must be dropped when config does not declare it")
	})

	t.Run("non-empty plan list survives empty config", func(t *testing.T) {
		plan := map[string]any{"id": "x", "tags": []any{"a", "b"}}
		config := map[string]any{"id": "x", "tags": []any{}}

		got := attr(t, encode(t, plan, config), "tags")
		var list []tftypes.Value
		require.NoError(t, got.As(&list))
		require.Len(t, list, 2)
	})

	t.Run("non-empty plan list survives missing config", func(t *testing.T) {
		plan := map[string]any{"id": "x", "tags": []any{"a"}}
		config := map[string]any{"id": "x"}

		got := attr(t, encode(t, plan, config), "tags")
		var list []tftypes.Value
		require.NoError(t, got.As(&list))
		require.Len(t, list, 1)
	})

	t.Run("scalar fields are untouched", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		config := map[string]any{"id": "x"}

		got := attr(t, encode(t, plan, config), "id")
		var s string
		require.NoError(t, got.As(&s))
		require.Equal(t, "x", s)
	})
}
