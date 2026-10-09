package mirrormakerreplicationflow

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// TestExpandConfigPropertiesExclude verifies the set is joined into the single
// string the API expects, that an empty string is sent when the value is being
// cleared, and that a field which was never configured is left out entirely: an
// empty string there would wipe the exclusions the API applies on its own.
func TestExpandConfigPropertiesExclude(t *testing.T) {
	t.Parallel()

	sch := &adapter.Schema{
		Type: adapter.SchemaTypeObject,
		Properties: map[string]*adapter.Schema{
			configPropertiesExclude: {
				Type:  adapter.SchemaTypeSet,
				Items: &adapter.Schema{Type: adapter.SchemaTypeString},
			},
		},
	}

	tests := []struct {
		name string
		plan map[string]any
		// state is nil for create, non-nil for update.
		state map[string]any
		// want is nil when the field must not be sent at all.
		want any
	}{
		{name: "multiple", plan: map[string]any{configPropertiesExclude: []any{"a", "b"}}, want: "a,b"},
		{name: "single", plan: map[string]any{configPropertiesExclude: []any{"a"}}, want: "a"},
		{name: "configured empty", plan: map[string]any{configPropertiesExclude: []any{}}, want: ""},
		{
			name:  "unchanged",
			plan:  map[string]any{configPropertiesExclude: []any{"a"}},
			state: map[string]any{configPropertiesExclude: []any{"a"}},
			want:  "a",
		},
		{
			name:  "removed",
			plan:  map[string]any{},
			state: map[string]any{configPropertiesExclude: []any{"a"}},
			want:  "",
		},
		{name: "never set, create", plan: map[string]any{}},
		{name: "never set, update", plan: map[string]any{}, state: map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := []adapter.ResourceDataOpt{adapter.WithTestPlan(tt.plan)}
			if tt.state != nil {
				opts = append(opts, adapter.WithTestState(tt.state))
			}

			d, err := adapter.NewResourceData(sch, nil, opts...)
			require.NoError(t, err)

			dto := make(map[string]any)
			require.NoError(t, expandConfigPropertiesExclude(d, dto))

			if tt.want == nil {
				require.NotContains(t, dto, configPropertiesExclude)
				return
			}
			require.Equal(t, tt.want, dto[configPropertiesExclude])
		})
	}
}

// TestFlattenConfigPropertiesExclude verifies the reverse, including that an
// empty string yields an empty set rather than a set holding one empty string.
func TestFlattenConfigPropertiesExclude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dto  map[string]any
		want map[string]any
	}{
		{
			name: "multiple",
			dto:  map[string]any{configPropertiesExclude: "a,b"},
			want: map[string]any{configPropertiesExclude: []any{"a", "b"}},
		},
		{
			name: "single",
			dto:  map[string]any{configPropertiesExclude: "a"},
			want: map[string]any{configPropertiesExclude: []any{"a"}},
		},
		{
			name: "empty string",
			dto:  map[string]any{configPropertiesExclude: ""},
			want: map[string]any{configPropertiesExclude: []any{}},
		},
		{
			name: "null is left alone",
			dto:  map[string]any{configPropertiesExclude: nil},
			want: map[string]any{configPropertiesExclude: nil},
		},
		{
			name: "missing is left alone",
			dto:  map[string]any{"other": "a,b"},
			want: map[string]any{"other": "a,b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.NoError(t, flattenConfigPropertiesExclude(nil, tt.dto))
			require.Equal(t, tt.want, tt.dto)
		})
	}
}

// TestFlattenConfigPropertiesExcludeUnexpectedType verifies that a non-string
// value is reported instead of being silently dropped.
func TestFlattenConfigPropertiesExcludeUnexpectedType(t *testing.T) {
	t.Parallel()

	dto := map[string]any{configPropertiesExclude: 42}
	err := flattenConfigPropertiesExclude(nil, dto)
	require.ErrorContains(t, err, "unexpected type int for config_properties_exclude")
}
