// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestResourceDataPrivateState(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("standalone data retains private values", func(t *testing.T) {
		d, err := NewResourceData(&Schema{Type: SchemaTypeObject}, nil)
		require.NoError(t, err)
		value, err := d.GetPrivate(ctx, "region")
		require.NoError(t, err)
		require.Nil(t, value)
		for _, want := range [][]byte{[]byte(`"eu-west-1"`), []byte(`""`), nil} {
			require.NoError(t, d.SetPrivate(ctx, "region", want))
			value, err = d.GetPrivate(ctx, "region")
			require.NoError(t, err)
			require.Equal(t, want, value)
		}
	})

	t.Run("response private state receives writes", func(t *testing.T) {
		private := &privateStateStub{values: map[string][]byte{"region": []byte(`"eu-west-1"`)}}
		d := &resourceData{private: private}
		require.NoError(t, d.SetPrivate(ctx, "region", []byte(`"us-east-1"`)))
		require.Equal(t, []byte(`"us-east-1"`), private.values["region"])
		value, err := d.GetPrivate(ctx, "region")
		require.NoError(t, err)
		require.Equal(t, private.values["region"], value)
	})

	t.Run("Framework diagnostics are returned as errors", func(t *testing.T) {
		private := &privateStateStub{diagnostics: diag.Diagnostics{diag.NewErrorDiagnostic("Private state error", "detail")}}
		d := &resourceData{private: private}
		_, err := d.GetPrivate(ctx, "region")
		require.EqualError(t, err, "Private state error: detail")
		require.EqualError(t, d.SetPrivate(ctx, "region", []byte(`"eu-west-1"`)), "Private state error: detail")
	})
}

type privateStateStub struct {
	values      map[string][]byte
	diagnostics diag.Diagnostics
}

func (p *privateStateStub) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return p.values[key], p.diagnostics
}

func (p *privateStateStub) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if p.diagnostics.HasError() {
		return p.diagnostics
	}
	p.values[key] = value
	return nil
}

func TestResourceDataRequiresReplace(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":       {Type: SchemaTypeString, Computed: true},
			"checksum": {Type: SchemaTypeString, Computed: true},
		},
	}

	d, err := NewResourceData(sch, []string{"id"})
	require.NoError(t, err)

	d.RequiresReplace("checksum", "checksum")
	require.Equal(t, []string{"checksum"}, d.requiresReplaceKeys())

	// A typo in the attribute name is a bug in the resource, not a user error.
	require.Panics(t, func() { d.RequiresReplace("nope") })
}

func TestResourceDataSetUnknown(t *testing.T) {
	t.Parallel()
	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":   {Type: SchemaTypeString, Computed: true},
			"tags": {Type: SchemaTypeMap, Items: &Schema{Type: SchemaTypeString}, Computed: true},
			"name": {Type: SchemaTypeString},
		},
	}
	t.Run("unknown values survive plan serialization", func(t *testing.T) {
		d, err := NewResourceData(sch, []string{"id"}, WithPreservePlanValues(), WithTestPlan(map[string]any{"id": "old"}))
		require.NoError(t, err)
		require.NoError(t, d.SetUnknown("id"))
		require.NoError(t, d.SetUnknown("tags"))
		var attributes map[string]tftypes.Value
		require.NoError(t, d.tfValue().As(&attributes))
		require.False(t, attributes["id"].IsKnown())
		require.True(t, attributes["id"].Type().Equal(tftypes.String))
		require.False(t, attributes["tags"].IsKnown())
		require.True(t, attributes["tags"].Type().Equal(tftypes.Map{ElementType: tftypes.String}))
		require.ErrorContains(t, d.SetUnknown("missing"), "not found in schema")
		require.ErrorContains(t, d.SetUnknown("name"), "not computed")
	})
	t.Run("unknown values cannot be written outside ModifyPlan", func(t *testing.T) {
		for _, opt := range []ResourceDataOpt{WithTestState(map[string]any{"id": "old"}), WithTestPlan(map[string]any{"id": "old"})} {
			d, err := NewResourceData(sch, []string{"id"}, opt)
			require.NoError(t, err)
			require.ErrorContains(t, d.SetUnknown("id"), "only supported during ModifyPlan")
			require.Equal(t, "old", d.ID())
		}
	})
}

// TestGetOkFalseIsAValue pins down that a stored false is a value, not a missing one.
// Read overrides that fill in a field the API never returns (jardeployment's
// restart_enabled) use GetOk to tell "the user set this" from "nothing is stored",
// and would overwrite an explicit false if the two looked the same.
func TestGetOkFalseIsAValue(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":              {Type: SchemaTypeString, Computed: true},
			"restart_enabled": {Type: SchemaTypeBool, Computed: true},
		},
	}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":              tftypes.String,
		"restart_enabled": tftypes.Bool,
	}}

	// Decodes through the same path as WithState, where a null attribute drops out
	// of the map while false stays in it.
	newData := func(t *testing.T, restartEnabled any) ResourceData {
		t.Helper()

		state, err := fromTFValue(sch, tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":              tftypes.NewValue(tftypes.String, "project/service"),
			"restart_enabled": tftypes.NewValue(tftypes.Bool, restartEnabled),
		}), false)
		require.NoError(t, err)

		d, err := NewResourceData(sch, []string{"id"}, WithTestState(state))
		require.NoError(t, err)
		return d
	}

	t.Run("false is stored", func(t *testing.T) {
		got, ok := newData(t, false).GetOk("restart_enabled")
		require.True(t, ok)
		require.Equal(t, false, got)
	})

	t.Run("null is missing", func(t *testing.T) {
		_, ok := newData(t, nil).GetOk("restart_enabled")
		require.False(t, ok)
	})
}

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

func TestNormalizeAnyPreservesEmptyMapStringValues(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"message": {Type: SchemaTypeString},
			"labels": {
				Type:  SchemaTypeMap,
				Items: &Schema{Type: SchemaTypeString},
			},
			"list": {
				Type:  SchemaTypeList,
				Items: &Schema{Type: SchemaTypeString},
			},
			"set": {
				Type:  SchemaTypeSet,
				Items: &Schema{Type: SchemaTypeString},
			},
		},
	}

	got, err := normalizeAny(sch, map[string]any{
		"message": "",
		"labels":  map[string]string{"empty": "", "value": "present"},
		"list":    []string{""},
		"set":     []string{""},
	}, false, true, false)
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"message": nil,
		"labels":  map[string]any{"empty": "", "value": "present"},
		"list":    []any{nil},
		"set":     []any{nil},
	}, got)
}

func TestResourceDataKeepsEmptyUserConfigStrings(t *testing.T) {
	t.Parallel()

	block := map[string]any{"site": "", "datadog_tags": []any{""}}
	check := func(t *testing.T, d ResourceData) {
		require.Nil(t, d.Get("endpoint_type"), "an empty string outside a user config is removed")
		require.Equal(t, []any{block}, d.Get("datadog_user_config"))
	}

	t.Run("set", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestState(map[string]any{}))
		require.NoError(t, err)
		require.NoError(t, d.Set("endpoint_type", ""))
		require.NoError(t, d.Set("datadog_user_config", block))
		check(t, d)
	})

	t.Run("flatten", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithIsDataSource(), WithTestState(map[string]any{}))
		require.NoError(t, err)
		require.NoError(t, d.Flatten(map[string]any{"endpoint_type": "", "datadog_user_config": block}))
		check(t, d)
	})
}

func TestResourceDataGetConfigOkDoesNotFallBack(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":    {Type: SchemaTypeString, Computed: true},
			"value": {Type: SchemaTypeString},
			"items": {Type: SchemaTypeSet, Items: &Schema{Type: SchemaTypeString}},
		},
	}
	rd, err := NewResourceData(
		sch,
		[]string{"id"},
		WithTestConfig(map[string]any{"value": "configured", "items": []string{}}),
		WithTestPlan(map[string]any{"value": "planned"}),
		WithTestState(map[string]any{"id": "example", "value": "stored"}),
	)
	require.NoError(t, err)

	value, ok := rd.GetConfigOk("value")
	require.True(t, ok)
	require.Equal(t, "configured", value)
	items, ok := rd.GetConfigOk("items")
	require.True(t, ok, "an explicitly configured empty collection must remain distinguishable from omission")
	require.Empty(t, items)
	_, ok = rd.GetConfigOk("id")
	require.False(t, ok)
}

func TestResourceDataGetConfigOkWithoutConfig(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"value": {Type: SchemaTypeString},
		},
	}
	rd, err := NewResourceData(
		sch,
		nil,
		WithTestState(map[string]any{"value": "stored"}),
	)
	require.NoError(t, err)

	_, ok := rd.GetConfigOk("value")
	require.False(t, ok, "state-only reads must not treat state as configuration")
}

func TestResourceDataGetStateOk(t *testing.T) {
	sch := &Schema{Type: SchemaTypeObject, Properties: map[string]*Schema{
		"id": {Type: SchemaTypeString, Computed: true},
	}}

	t.Run("create with checkpointed ID has no prior state", func(t *testing.T) {
		d, err := NewResourceData(sch, []string{"id"}, WithTestPlan(map[string]any{}))
		require.NoError(t, err)
		require.NoError(t, d.SetID("new-id"))
		require.Equal(t, "new-id", d.ID())
		_, ok := d.GetStateOk("id")
		require.False(t, ok)
	})

	t.Run("update retains prior ID after migration", func(t *testing.T) {
		d, err := NewResourceData(sch, []string{"id"},
			WithTestState(map[string]any{"id": "legacy-id"}),
			WithTestPlan(map[string]any{"id": "new-id"}),
		)
		require.NoError(t, err)
		id, ok := d.GetStateOk("id")
		require.True(t, ok)
		require.Equal(t, "legacy-id", id)
		require.Equal(t, "legacy-id", d.GetState("id"))
	})
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

// TestResourceDataHasChangeNestedNulls ModifyPlan keeps the nulls of the plan, which the state has none of,
// e.g., {"foo": "a", "bar": null} is {"foo": "a"} in the state.
func TestResourceDataHasChangeNestedNulls(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id": {Type: SchemaTypeString, Computed: true},
			"config": {
				Type:     SchemaTypeList,
				IsObject: true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"foo": {Type: SchemaTypeString},
						"bar": {Type: SchemaTypeString},
					},
				},
			},
		},
	}

	cases := []struct {
		name string
		bar  any
		want bool
	}{
		{name: "null", bar: tftypes.NewValue(tftypes.String, nil), want: false},
		{name: "unknown", bar: tftypes.NewValue(tftypes.String, tftypes.UnknownValue), want: true},
		{name: "value", bar: "b", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rd, err := NewResourceData(sch, []string{"id"},
				WithTestPlan(map[string]any{"id": "x", "config": []any{map[string]any{"foo": "a", "bar": tc.bar}}}),
				WithTestState(map[string]any{"id": "x", "config": []any{map[string]any{"foo": "a"}}}),
			)
			require.NoError(t, err)
			require.Equal(t, tc.want, rd.HasChange("config"))
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

func TestResourceDataFlattenOptionalIDFields(t *testing.T) {
	t.Parallel()
	sch := &Schema{Type: SchemaTypeObject, Properties: map[string]*Schema{
		"id":      {Type: SchemaTypeString, Computed: true},
		"project": {Type: SchemaTypeString},
		"name":    {Type: SchemaTypeString},
		"region":  {Type: SchemaTypeString},
	}}
	newData := func(t *testing.T) ResourceData {
		t.Helper()
		d, err := NewResourceData(sch, []string{"project", "name", "region"},
			WithIdentityFieldsOptional([]string{"name", "region"}), WithTestState(map[string]any{}))
		require.NoError(t, err)
		return d
	}

	t.Run("all fields present", func(t *testing.T) {
		d := newData(t)
		require.NoError(t, d.Flatten(map[string]any{"project": "project", "name": "example", "region": "region"}))
		require.Equal(t, "project/example/region", d.ID())
	})

	t.Run("optional suffix omitted", func(t *testing.T) {
		d := newData(t)
		require.NoError(t, d.Flatten(map[string]any{"project": "project", "name": "example"}))
		require.Equal(t, "project/example", d.ID())
	})

	t.Run("all optional fields omitted", func(t *testing.T) {
		d := newData(t)
		require.NoError(t, d.Flatten(map[string]any{"project": "project"}))
		require.Equal(t, "project", d.ID())
	})

	t.Run("middle field omitted", func(t *testing.T) {
		d := newData(t)
		err := d.Flatten(map[string]any{"project": "project", "region": "region"})
		require.ErrorContains(t, err, "preceding optional field")
		require.Empty(t, d.ID())
	})

	t.Run("required field omitted", func(t *testing.T) {
		d := newData(t)
		require.ErrorContains(t, d.Flatten(map[string]any{"name": "example"}), `no value for id field "project"`)
		require.Empty(t, d.ID())
	})
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

// TestResourceDataFlattenComputedAttributes verifies that API-owned nested attributes
// preserve explicit empty collections and remove prior state only when omitted.
func TestResourceDataFlattenComputedAttributes(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":   {Type: SchemaTypeString, Computed: true},
			"name": {Type: SchemaTypeString},
			"versions": {
				Type:        SchemaTypeList,
				Computed:    true,
				IsAttribute: true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"id": {Type: SchemaTypeString, Computed: true},
					},
				},
			},
			"deployment": {
				Type:        SchemaTypeList,
				Computed:    true,
				IsAttribute: true,
				IsObject:    true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"id": {Type: SchemaTypeString, Computed: true},
					},
				},
			},
		},
	}

	stateValue := func(t *testing.T, rd ResourceData, name string) tftypes.Value {
		t.Helper()

		var attrs map[string]tftypes.Value
		require.NoError(t, rd.tfValue().As(&attrs))
		return attrs[name]
	}

	t.Run("explicit empty list stays empty", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{
			"name":     "my-resource",
			"versions": []any{},
		}))

		got, ok := rd.GetOk("versions")
		require.True(t, ok)
		require.Equal(t, []any{}, got)
		require.False(t, stateValue(t, rd, "versions").IsNull())
	})

	t.Run("omitted object removes stale state", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{
				"name":       "my-resource",
				"deployment": []any{map[string]any{"id": "stale"}},
			}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{"name": "my-resource"}))

		_, ok := rd.GetOk("deployment")
		require.False(t, ok)
		require.True(t, stateValue(t, rd, "deployment").IsNull())
	})
}

// TestResourceDataEmptyCollections verifies that Flatten keeps an empty collection apart
// from a null one: the API reports both as "no values", but Terraform requires the state
// to keep the shape the configuration asked for.
func TestResourceDataEmptyCollections(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":   {Type: SchemaTypeString, Computed: true},
			"name": {Type: SchemaTypeString},
			"keys": {Type: SchemaTypeList, Items: &Schema{Type: SchemaTypeString}},
			"tags": {Type: SchemaTypeMap, Items: &Schema{Type: SchemaTypeString}},
		},
	}

	// The API has no values for "keys" and "tags".
	flattenIn := map[string]any{
		"name": "my-resource",
		"keys": []any{},
		"tags": map[string]any{},
	}

	// stateValue returns the value written to the Terraform state.
	stateValue := func(t *testing.T, rd ResourceData, name string) tftypes.Value {
		t.Helper()

		var attrs map[string]tftypes.Value
		require.NoError(t, rd.tfValue().As(&attrs))
		return attrs[name]
	}

	// flatten applies the API values on top of a plan that mirrors the configuration.
	flatten := func(t *testing.T, plan map[string]any) ResourceData {
		t.Helper()

		rd, err := NewResourceData(sch, []string{"name"},
			WithTestPlan(plan),
			WithTestConfig(plan),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(flattenIn))
		return rd
	}

	t.Run("configured empty list stays empty", func(t *testing.T) {
		rd := flatten(t, map[string]any{"name": "my-resource", "keys": []any{}})

		got, ok := rd.GetOk("keys")
		require.True(t, ok)
		require.Equal(t, []any{}, got)
		require.False(t, stateValue(t, rd, "keys").IsNull(), "a configured empty list must not become null")
	})

	t.Run("configured empty map stays empty", func(t *testing.T) {
		rd := flatten(t, map[string]any{"name": "my-resource", "tags": map[string]any{}})

		got, ok := rd.GetOk("tags")
		require.True(t, ok)
		require.Equal(t, map[string]any{}, got)
		require.False(t, stateValue(t, rd, "tags").IsNull(), "a configured empty map must not become null")
	})

	t.Run("unset collections stay null", func(t *testing.T) {
		rd := flatten(t, map[string]any{"name": "my-resource"})

		_, ok := rd.GetOk("keys")
		require.False(t, ok)
		require.True(t, stateValue(t, rd, "keys").IsNull(), "an unset list must not become an empty list")

		_, ok = rd.GetOk("tags")
		require.False(t, ok)
		require.True(t, stateValue(t, rd, "tags").IsNull(), "an unset map must not become an empty map")
	})

	t.Run("empty collections survive a refresh", func(t *testing.T) {
		// Read gets no config, so the prior state is the only sign that the empty
		// collections were configured. Dropping them here would plan a null to "[]" change
		// on the next run.
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{
				"name": "my-resource",
				"keys": []any{},
				"tags": map[string]any{},
			}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(flattenIn))

		require.False(t, stateValue(t, rd, "keys").IsNull(), "an empty list in state must not become null")
		require.False(t, stateValue(t, rd, "tags").IsNull(), "an empty map in state must not become null")
	})

	t.Run("values dropped by the API are reported as empty", func(t *testing.T) {
		rd, err := NewResourceData(sch, []string{"name"},
			WithTestState(map[string]any{
				"name": "my-resource",
				"keys": []any{"stale"},
				"tags": map[string]any{"stale": "value"},
			}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(flattenIn))

		require.Equal(t, []any{}, rd.Get("keys"))
		require.Equal(t, map[string]any{}, rd.Get("tags"))
	})

	t.Run("computed optional set kept when unset", func(t *testing.T) {
		computedSch := &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"id":         {Type: SchemaTypeString, Computed: true},
				"name":       {Type: SchemaTypeString},
				"static_ips": {Type: SchemaTypeSet, Computed: true, IsAttribute: true, Items: &Schema{Type: SchemaTypeString}},
			},
		}
		rd, err := NewResourceData(computedSch, []string{"name"},
			WithTestPlan(map[string]any{"name": "my-resource"}),
			WithTestConfig(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{"name": "my-resource", "static_ips": []any{}}))

		got, ok := rd.GetOk("static_ips")
		require.True(t, ok)
		require.Equal(t, []any{}, got)
		require.False(t, stateValue(t, rd, "static_ips").IsNull())
	})

	t.Run("optional set pruned when unset", func(t *testing.T) {
		optionalSch := &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"id":        {Type: SchemaTypeString, Computed: true},
				"name":      {Type: SchemaTypeString},
				"allowlist": {Type: SchemaTypeSet, Items: &Schema{Type: SchemaTypeString}},
			},
		}
		rd, err := NewResourceData(optionalSch, []string{"name"},
			WithTestPlan(map[string]any{"name": "my-resource"}),
			WithTestConfig(map[string]any{"name": "my-resource"}),
		)
		require.NoError(t, err)
		require.NoError(t, rd.Flatten(map[string]any{"name": "my-resource", "allowlist": []any{}}))

		_, ok := rd.GetOk("allowlist")
		require.False(t, ok)
		require.True(t, stateValue(t, rd, "allowlist").IsNull())
	})
}

func TestHasConfiguredChange(t *testing.T) {
	t.Parallel()

	userBlock := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"backup_hour": {Type: SchemaTypeInt},
			"pg_version":  {Type: SchemaTypeString},
			"ip_filter":   {Type: SchemaTypeSet, Items: &Schema{Type: SchemaTypeString}},
			"pglookout": {
				Type:     SchemaTypeList,
				IsObject: true,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"max_failover_replication_time_lag": {Type: SchemaTypeInt},
					},
				},
			},
		},
	}

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":              {Type: SchemaTypeString, Computed: true},
			"foo_user_config": {Type: SchemaTypeList, IsObject: true, Items: userBlock},
		},
	}

	fullStateBlock := map[string]any{
		"backup_hour": 16,
		"pg_version":  "16",
		"ip_filter":   []any{"0.0.0.0/0"},
		"pglookout":   []any{map[string]any{"max_failover_replication_time_lag": 60}},
	}

	cases := []struct {
		name   string
		config map[string]any
		state  map[string]any
		want   bool
	}{
		{
			name:   "config omits user config block",
			config: map[string]any{"id": "svc"},
			state:  map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:   false,
		},
		{
			name: "configured option matches state",
			config: map[string]any{
				"id":              "svc",
				"foo_user_config": []any{map[string]any{"backup_hour": 16}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  false,
		},
		{
			name: "configured int differs",
			config: map[string]any{
				"id":              "svc",
				"foo_user_config": []any{map[string]any{"backup_hour": 10}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{map[string]any{"backup_hour": 16}}},
			want:  true,
		},
		{
			name: "configured int unknown",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"backup_hour": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{map[string]any{"backup_hour": 16}}},
			want:  true,
		},
		{
			name: "null option is not a configured change",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"pg_version": tftypes.NewValue(tftypes.String, nil),
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{map[string]any{"pg_version": "16"}}},
			want:  false,
		},
		{
			name: "configured set differs",
			config: map[string]any{
				"id":              "svc",
				"foo_user_config": []any{map[string]any{"ip_filter": []any{"10.0.0.0/8"}}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{map[string]any{"ip_filter": []any{"0.0.0.0/0"}}}},
			want:  true,
		},
		{
			name: "empty nested block is not a change",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"pglookout": []any{map[string]any{
						"max_failover_replication_time_lag": tftypes.NewValue(tftypes.Number, nil),
					}},
				}},
			},
			state: map[string]any{
				"id":              "svc",
				"foo_user_config": []any{map[string]any{"backup_hour": 16}},
			},
			want: false,
		},
		{
			name: "nested block option differs",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"pglookout": []any{map[string]any{"max_failover_replication_time_lag": 30}},
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  true,
		},
		{
			// The framework passes an absent block as null, see NullifyCollectionBlocks
			name:   "null user config block is not a change",
			config: map[string]any{"id": "svc", "foo_user_config": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{}}, nil)},
			state:  map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:   false,
		},
		{
			name: "null nested block is not a change",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"backup_hour": 16,
					"pglookout":   tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{}}, nil),
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  false,
		},
		{
			name: "emptied set option is a change",
			config: map[string]any{
				"id":              "svc",
				"foo_user_config": []any{map[string]any{"ip_filter": []any{}}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  true,
		},
		{
			name: "config has more blocks than the state",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"pglookout": []any{
						map[string]any{"max_failover_replication_time_lag": 60},
						map[string]any{"max_failover_replication_time_lag": 30},
					},
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  true,
		},
		{
			name: "block with only null placeholders",
			config: map[string]any{
				"id": "svc",
				"foo_user_config": []any{map[string]any{
					"backup_hour": tftypes.NewValue(tftypes.Number, nil),
					"pg_version":  tftypes.NewValue(tftypes.String, nil),
					"ip_filter":   tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
					"pglookout": []any{map[string]any{
						"max_failover_replication_time_lag": tftypes.NewValue(tftypes.Number, nil),
					}},
				}},
			},
			state: map[string]any{"id": "svc", "foo_user_config": []any{fullStateBlock}},
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rd, err := NewResourceData(sch, []string{"id"},
				WithTestState(tc.state),
				WithTestConfig(tc.config),
			)
			require.NoError(t, err)
			require.Equal(t, tc.want, rd.HasConfiguredChange("foo_user_config"))
		})
	}
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

func TestResourceDataExpandWithoutChanges(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":          {Type: SchemaTypeString, Computed: true},
			"name":        {Type: SchemaTypeString},
			"enabled":     {Type: SchemaTypeBool},
			"description": {Type: SchemaTypeString, Computed: true},
			"password_wo": {Type: SchemaTypeString},
			"version":     {Type: SchemaTypeInt},
			"config": {
				Type:     SchemaTypeList,
				IsObject: true,
				Computed: true,
				Items: &Schema{
					Type:       SchemaTypeObject,
					Properties: map[string]*Schema{"retention": {Type: SchemaTypeString, Computed: true}},
				},
			},
			"timeouts": {
				Type:       SchemaTypeObject,
				Properties: map[string]*Schema{"update": {Type: SchemaTypeString}},
			},
		},
	}
	type request struct {
		Name        string         `json:"name,omitempty"`
		Enabled     bool           `json:"enabled,omitempty"`
		Description *string        `json:"description,omitempty"`
		Password    *string        `json:"password,omitempty"`
		Config      map[string]any `json:"config,omitempty"`
	}
	state := func() map[string]any {
		return map[string]any{
			"id":          "foo",
			"name":        "foo",
			"enabled":     true,
			"description": "bar",
			"version":     1,
			"config":      []any{map[string]any{"retention": "1"}},
		}
	}
	// Like a lookup of a deleted billing group card
	lookupName := func(d ResourceData, dto map[string]any) error {
		if d.Get("name") == "foo" {
			return errors.New("name not found")
		}
		return nil
	}
	writeOnly := func(d ResourceData, dto map[string]any) error {
		if v, ok := d.GetOk("password_wo"); ok {
			dto["password"] = v
		}
		return nil
	}

	cases := []struct {
		name      string
		plan      func(map[string]any)
		config    map[string]any
		modifiers []MapModifier
		changed   bool
	}{
		{name: "same values"},
		{name: "timeouts", plan: func(m map[string]any) { m["timeouts"] = map[string]any{"update": "20m"} }},
		{name: "computed field unknown until apply", plan: func(m map[string]any) { delete(m, "description") }},
		{name: "removed computed block", plan: func(m map[string]any) { delete(m, "config") }},
		{name: "changed computed block", plan: func(m map[string]any) {
			m["config"] = []any{map[string]any{"retention": "2"}}
		}, changed: true},
		{name: "changed field", plan: func(m map[string]any) { m["name"] = "baz" }, changed: true},
		{name: "zero value", plan: func(m map[string]any) { m["enabled"] = false }, changed: true},
		{name: "removed field", plan: func(m map[string]any) { delete(m, "name") }, changed: true},
		{name: "field outside the request", plan: func(m map[string]any) { m["version"] = 2 }, changed: true},
		{name: "write-only value without its version", config: map[string]any{"password_wo": "secret"}, modifiers: []MapModifier{writeOnly}},
		{name: "modifiers don't run for the state", plan: func(m map[string]any) { m["name"] = "baz" }, modifiers: []MapModifier{lookupName}, changed: true},
		{name: "modifiers don't run without changes", modifiers: []MapModifier{lookupName}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			plan := state()
			if tc.plan != nil {
				tc.plan(plan)
			}
			d, err := NewResourceData(sch, nil, WithTestPlan(plan), WithTestState(state()), WithTestConfig(tc.config))
			require.NoError(t, err)

			err = d.Expand(new(request), tc.modifiers...)
			if tc.changed {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, ErrNoChanges)
		})
	}

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithTestPlan(state()))
		require.NoError(t, err)
		require.NoError(t, d.Expand(new(request)))
	})
}

func TestResourceDataHoistedFields(t *testing.T) {
	t.Parallel()

	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"pg_allow_replication": {Type: SchemaTypeBool, HoistedFrom: "access_control"},
			"valkey_acl_keys":      {Type: SchemaTypeList, Items: &Schema{Type: SchemaTypeString}, HoistedFrom: "access_control"},
			"username":             {Type: SchemaTypeString},
			"id":                   {Type: SchemaTypeString, Computed: true},
		},
	}

	t.Run("expand nests the fields after the modifiers", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithTestPlan(map[string]any{
			"username":        "u",
			"valkey_acl_keys": []any{"k"},
		}))
		require.NoError(t, err)
		setReplication := func(_ ResourceData, dto map[string]any) error {
			dto["pg_allow_replication"] = true
			return nil
		}
		var out map[string]any
		require.NoError(t, d.Expand(&out, setReplication))
		require.Equal(t, map[string]any{
			"username": "u",
			"access_control": map[string]any{
				"pg_allow_replication": true,
				"valkey_acl_keys":      []any{"k"},
			},
		}, out)
	})

	t.Run("expand sends no object without the fields", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithTestPlan(map[string]any{"username": "u"}))
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, d.Expand(&out))
		require.Equal(t, map[string]any{"username": "u"}, out)
	})

	t.Run("flatten lifts the fields before the modifiers", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, []string{"username"}, WithTestState(map[string]any{}))
		require.NoError(t, err)
		rsp := map[string]any{
			"username": "u",
			"access_control": map[string]any{
				"valkey_acl_keys": []any{"k"},
				"redis_acl_keys":  []any{"r"},
				"username":        "nested",
			},
		}
		var seen map[string]any
		record := func(_ ResourceData, dto map[string]any) error {
			seen = maps.Clone(dto)
			return nil
		}
		require.NoError(t, d.Flatten(rsp, record))
		require.Equal(t, map[string]any{"username": "u", "valkey_acl_keys": []any{"k"}}, seen)
		require.Equal(t, "u", d.Get("username"))
		require.Equal(t, []any{"k"}, d.Get("valkey_acl_keys"))
	})

	t.Run("flatten keeps the state of an omitted field", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, []string{"username"}, WithTestState(map[string]any{"username": "u", "valkey_acl_keys": []any{"k"}}))
		require.NoError(t, err)
		require.NoError(t, d.Flatten(map[string]any{"username": "u", "access_control": map[string]any{}}))
		require.Equal(t, []any{"k"}, d.Get("valkey_acl_keys"))
	})

	t.Run("has change hoisted", func(t *testing.T) {
		t.Parallel()
		state := map[string]any{"username": "u", "pg_allow_replication": false}
		changed, err := NewResourceData(sch, nil, WithTestState(state), WithTestPlan(map[string]any{"username": "u", "pg_allow_replication": true}))
		require.NoError(t, err)
		require.True(t, changed.HasChangeHoisted("access_control"))
		require.False(t, changed.HasChangeHoisted("other"))

		rootOnly, err := NewResourceData(sch, nil, WithTestState(state), WithTestPlan(map[string]any{"username": "v", "pg_allow_replication": false}))
		require.NoError(t, err)
		require.False(t, rootOnly.HasChangeHoisted("access_control"))
	})
}
