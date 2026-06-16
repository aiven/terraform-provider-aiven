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

	rd, err := NewResourceData(sch, nil,
		WithTestPlan(map[string]any{
			"nested": map[string]string{"bar": "baz"},
			"tags":   []string{"a", "b"},
			"labels": map[string]string{"k": "v"},
		}),
	)
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
			rd, err := NewResourceData(sch, []string{"id"},
				WithTestPlan(tc.plan),
			)
			require.NoError(t, err)

			var got request
			require.NoError(t, rd.Expand(&got))
			require.Equal(t, tc.want, got)

			require.NotPanics(t, func() { _ = rd.(*resourceData).tfValue() })
		})
	}
}

// TestResourceDataGetOkSplitsCompositeID verifies that GetOk derives id-field
// component values by splitting the composite "id" string from state when the
// components are not stored as their own top-level fields (e.g. legacy state
// produced by older provider versions, or right after import).
func TestResourceDataGetOkSplitsCompositeID(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":      {Type: SchemaTypeString, Computed: true},
			"project": {Type: SchemaTypeString},
			"vpc_id":  {Type: SchemaTypeString},
		},
	}
	idFields := []string{"project", "vpc_id"}

	cases := []struct {
		name        string
		plan        map[string]any
		state       map[string]any
		wantProject string
		wantVPC     string
	}{
		{
			// Read/Delete flow: only state, and state has only the composite id.
			name:        "no_plan_components_derived_from_composite_id",
			state:       map[string]any{"id": "my-proj/vpc-123"},
			wantProject: "my-proj",
			wantVPC:     "vpc-123",
		},
		{
			// Update flow: plan is set but does not carry id components;
			// they must still be derived from state's composite id for API calls.
			name:        "update_with_plan_still_derives_id_components",
			plan:        map[string]any{},
			state:       map[string]any{"id": "my-proj/vpc-123"},
			wantProject: "my-proj",
			wantVPC:     "vpc-123",
		},
		{
			// When the component is stored directly in state, that value wins
			// over splitting the (potentially stale) composite id.
			name: "direct_state_field_takes_precedence_over_split",
			state: map[string]any{
				"id":      "stale/old",
				"project": "fresh-proj",
				"vpc_id":  "fresh-vpc",
			},
			wantProject: "fresh-proj",
			wantVPC:     "fresh-vpc",
		},
		{
			// Some resources allow changing id-field components on Update
			// (the new values arrive via the plan). The plan is checked before
			// state so those new values are picked up — otherwise we'd send
			// stale ids from state to the API.
			name: "plan_takes_precedence_over_state",
			plan: map[string]any{
				"project": "plan-proj",
				"vpc_id":  "plan-vpc",
			},
			state: map[string]any{
				"id":      "state-proj/state-vpc",
				"project": "state-proj",
				"vpc_id":  "state-vpc",
			},
			wantProject: "plan-proj",
			wantVPC:     "plan-vpc",
		},
		{
			// No state at all: nothing to fall back to.
			name:        "no_state_returns_zero",
			wantProject: "",
			wantVPC:     "",
		},
		{
			// State has neither the components nor a composite id.
			name:        "empty_state_returns_zero",
			state:       map[string]any{},
			wantProject: "",
			wantVPC:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rd, err := NewResourceData(sch, idFields,
				WithTestPlan(tc.plan),
				WithTestState(tc.state),
			)
			require.NoError(t, err)

			require.Equal(t, tc.wantProject, rd.Get("project"), "project")
			require.Equal(t, tc.wantVPC, rd.Get("vpc_id"), "vpc_id")
		})
	}
}

// TestResourceDataFlattenRemovesEmptyComputedBlocks verifies that Flatten strips
// unconfigured computed object blocks from resources but keeps them for data sources.
func TestResourceDataFlattenRemovesEmptyComputedBlocks(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":   {Type: SchemaTypeString, Computed: true},
			"name": {Type: SchemaTypeString},
			"metadata": {
				Type:     SchemaTypeList,
				IsObject: true,
				Computed: true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"key": {Type: SchemaTypeString, Computed: true},
					},
				},
			},
		},
	}
	flattenIn := map[string]any{
		"name": "my-resource",
		"metadata": []any{
			map[string]any{"key": "value"},
		},
	}

	t.Run("resource omits unconfigured computed block", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{"name": "my-resource"}),
			WithTestConfig(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(flattenIn))

		_, ok := rd.GetOk("metadata")
		require.False(t, ok)
	})

	t.Run("resource removes stale unconfigured computed block", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{
				"name":     "my-resource",
				"metadata": []any{map[string]any{"key": "stale"}},
			}),
			WithTestConfig(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{"name": "my-resource"}))

		_, ok := rd.GetOk("metadata")
		require.False(t, ok)
	})

	t.Run("resource keeps configured computed block", func(t *testing.T) {
		configMetadata := []any{
			map[string]any{"key": "configured"},
		}

		rd, err := NewResourceData(sch, []string{"name"},
			WithTestPlan(map[string]any{
				"name":     "my-resource",
				"metadata": configMetadata,
			}),
			WithTestConfig(map[string]any{
				"name":     "my-resource",
				"metadata": configMetadata,
			}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{
			"name":     "my-resource",
			"metadata": configMetadata,
		}))

		got, ok := rd.GetOk("metadata")
		require.True(t, ok)
		require.Equal(t, configMetadata, got)
	})

	t.Run("data source keeps computed block", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithIsDataSource(),
			WithTestState(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(flattenIn))

		got, ok := rd.GetOk("metadata")
		require.True(t, ok)
		require.Equal(t, flattenIn["metadata"], got)
	})
}

func TestResourceDataSingletonObjectBlocks(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"name": {Type: SchemaTypeString},
			"config": {
				Type:     SchemaTypeList,
				IsObject: true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"retention_ms": {Type: SchemaTypeString},
					},
				},
			},
		},
	}

	expand := func(t *testing.T, plan map[string]any) (map[string]any, error) {
		t.Helper()

		rd, err := NewResourceData(sch, []string{"name"}, WithTestPlan(plan))
		require.NoError(t, err)

		var out map[string]any
		err = rd.Expand(&out)
		return out, err
	}

	t.Run("allows missing object block", func(t *testing.T) {
		got, err := expand(t, map[string]any{"name": "topic"})

		require.NoError(t, err)
		require.Equal(t, map[string]any{"name": "topic"}, got)
	})

	t.Run("allows one object block", func(t *testing.T) {
		got, err := expand(t, map[string]any{
			"name": "topic",
			"config": []any{
				map[string]any{"retention_ms": "1000"},
			},
		})

		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"name": "topic",
			"config": map[string]any{
				"retention_ms": "1000",
			},
		}, got)
	})

	t.Run("rejects multiple object blocks", func(t *testing.T) {
		_, err := expand(t, map[string]any{
			"name": "topic",
			"config": []any{
				map[string]any{"retention_ms": "1000"},
				map[string]any{"retention_ms": "2000"},
			},
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "expected at most one object, got 2")
	})
}
