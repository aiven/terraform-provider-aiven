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

// TestFlatten_EmptyCollectionOverlay covers the Flatten rule that drops
// empty-collection fields from an API response when the user (plan in
// Create/Update, prior state in Read) never set them. Without the rule,
// overlaying produces encoded `SetValEmpty` for attrs that were null in
// plan, which TF rejects as "inconsistent result after apply".
//
// Symmetric to TestToTFValue_EmptyCollectionsAreKnown (marshalling_test.go):
// that test covers the Required-empty encode; this one covers the
// Optional-omitted filter so both cases work together.
func TestFlatten_EmptyCollectionOverlay(t *testing.T) {
	t.Parallel()

	strSch := &Schema{Type: SchemaTypeString}
	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":     strSch,
			"tags":   {Type: SchemaTypeList, Items: strSch},
			"labels": {Type: SchemaTypeMap, Items: strSch},
		},
	}

	attr := func(t *testing.T, v tftypes.Value, name string) tftypes.Value {
		t.Helper()
		var m map[string]tftypes.Value
		require.NoError(t, v.As(&m))
		got, ok := m[name]
		require.True(t, ok, "attribute %q missing", name)
		return got
	}

	t.Run("API empty, plan omitted → dropped, encodes null", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, plan, nil, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":   "x",
			"tags": []any{},
		}))

		v := rd.(*resourceData).tfValue()
		require.True(t, attr(t, v, "tags").IsNull())
	})

	t.Run("API empty, plan has explicit [] → preserved, encodes empty-known", func(t *testing.T) {
		plan := map[string]any{"id": "x", "tags": []any{}}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, plan, nil, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":   "x",
			"tags": []any{},
		}))

		v := rd.(*resourceData).tfValue()
		got := attr(t, v, "tags")
		require.False(t, got.IsNull())
		require.True(t, got.IsKnown())
	})

	t.Run("Read path: API empty, prior state omitted → dropped", func(t *testing.T) {
		state := map[string]any{"id": "x"}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, nil, state, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":   "x",
			"tags": []any{},
		}))

		v := rd.(*resourceData).tfValue()
		require.True(t, attr(t, v, "tags").IsNull())
	})

	t.Run("Read path: API empty, prior state has [] → preserved", func(t *testing.T) {
		state := map[string]any{"id": "x", "tags": []any{}}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, nil, state, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":   "x",
			"tags": []any{},
		}))

		v := rd.(*resourceData).tfValue()
		require.False(t, attr(t, v, "tags").IsNull())
	})

	t.Run("empty map follows same rule", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, plan, nil, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":     "x",
			"labels": map[string]any{},
		}))

		v := rd.(*resourceData).tfValue()
		require.True(t, attr(t, v, "labels").IsNull())
	})

	t.Run("non-empty API response always preserved", func(t *testing.T) {
		plan := map[string]any{"id": "x"}
		rd, err := NewResourceDataFromMaps(sch, []string{"id"}, plan, nil, nil)
		require.NoError(t, err)

		require.NoError(t, rd.(*resourceData).Flatten(map[string]any{
			"id":   "x",
			"tags": []any{"a"},
		}))

		v := rd.(*resourceData).tfValue()
		got := attr(t, v, "tags")
		var list []tftypes.Value
		require.NoError(t, got.As(&list))
		require.Len(t, list, 1)
	})
}
