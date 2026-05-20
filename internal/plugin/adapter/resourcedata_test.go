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

// TestResourceDataAcceptsTypedContainers verifies that Expand, tfValue and
// Set all transparently accept strictly-typed string-keyed maps and slices
// (e.g. map[string]string, []string) via reflective conversion in
// asAnyMap / asAnySlice, with no JSON round-trip.
func TestResourceDataAcceptsTypedContainers(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"nested": {
				Type: SchemaTypeObject,
				Properties: map[string]*Schema{
					"bar": {Type: SchemaTypeString},
				},
			},
			"tags": {
				Type:  SchemaTypeList,
				Items: &Schema{Type: SchemaTypeString},
			},
			"labels": {
				Type:  SchemaTypeMap,
				Items: &Schema{Type: SchemaTypeString},
			},
		},
	}

	rd, err := NewResourceDataFromMaps(sch, nil, map[string]any{
		"nested": map[string]string{"bar": "baz"},
		"tags":   []string{"a", "b"},
		"labels": map[string]string{"k": "v"},
	}, nil, nil)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, rd.Expand(&out))
	require.Equal(t, map[string]any{
		"nested": map[string]any{"bar": "baz"},
		"tags":   []any{"a", "b"},
		"labels": map[string]any{"k": "v"},
	}, out)

	require.NotPanics(t, func() { _ = rd.(*resourceData).tfValue() })

	require.NoError(t, rd.Set("nested", map[string]string{"bar": "baz2"}))
	require.NoError(t, rd.Set("tags", []string{"c"}))
	require.NoError(t, rd.Set("labels", map[string]string{"k": "v2"}))
}

// TestResourceDataPreservedUnknownAndNullPlan verifies that preserved
// tftypes.Value entries (produced by fromTFValue with preservePlanValues=true
// for the ModifyPlan flow) are coerced to zero values by normalizeAny, so
// Expand and tfValue never see the raw tftypes.Value and never panic. This
// mirrors the existing tftypes.Value short-circuit in getOk.
func TestResourceDataPreservedUnknownAndNullPlan(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":      {Type: SchemaTypeString, Computed: true},
			"name":    {Type: SchemaTypeString},
			"enabled": {Type: SchemaTypeBool},
		},
	}

	type request struct {
		Enabled bool   `json:"enabled,omitempty"`
		Name    string `json:"name,omitempty"`
	}

	cases := []struct {
		name string
		plan map[string]any
		want request
	}{
		{
			name: "unknown_string",
			plan: map[string]any{
				"name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			},
		},
		{
			name: "null_bool",
			plan: map[string]any{
				"name":    "target",
				"enabled": tftypes.NewValue(tftypes.Bool, nil),
			},
			want: request{Name: "target"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rd, err := NewResourceDataFromMaps(sch, []string{"id"}, tc.plan, nil, nil)
			require.NoError(t, err)

			var got request
			require.NoError(t, rd.Expand(&got))
			require.Equal(t, tc.want, got)

			require.NotPanics(t, func() { _ = rd.(*resourceData).tfValue() })
		})
	}
}
