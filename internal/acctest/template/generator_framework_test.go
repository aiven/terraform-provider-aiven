package template

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFrameworkGenerateTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		schema       interface{}
		resourceType string
		kind         ResourceKind
		want         string
	}{
		{
			name: "resource with boolean fields",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"enable_feature": resourceschema.BoolAttribute{
						Required: true,
					},
					"disable_feature": resourceschema.BoolAttribute{
						Optional: true,
					},
					"settings": resourceschema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]resourceschema.Attribute{
							"feature_one": resourceschema.BoolAttribute{
								Required: true,
							},
							"feature_two": resourceschema.BoolAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  {{- if ne .disable_feature nil }}
  disable_feature = {{ .disable_feature }}
  {{- end }}
  enable_feature = {{ .enable_feature }}
  settings {
    feature_one = {{ (index .settings "feature_one") }}
    {{- if ne (index .settings "feature_two") nil }}
    feature_two = {{ (index .settings "feature_two") }}
    {{- end }}
  }
}`,
		},
		{
			name: "basic resource with required string field",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"name": resourceschema.StringAttribute{
						Required: true,
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  name = {{ renderValue (required .name) }}
}`,
		},
		{
			name: "resource with optional fields",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"name": resourceschema.StringAttribute{
						Required: true,
					},
					"description": resourceschema.StringAttribute{
						Optional: true,
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  {{- if .description }}
  description = {{ renderValue .description }}
  {{- end }}
  name = {{ renderValue (required .name) }}
}`,
		},
		{
			name: "data source with required fields",
			schema: datasourceschema.Schema{
				Attributes: map[string]datasourceschema.Attribute{
					"id": datasourceschema.StringAttribute{
						Required: true,
					},
				},
			},
			resourceType: "test_datasource",
			kind:         ResourceKindDataSource,
			want: `data "test_datasource" "{{ required .resource_name }}" {
  id = {{ renderValue (required .id) }}
}`,
		},
		{
			name: "resource with map field",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"tags": resourceschema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  {{- if .tags }}
  tags = {
    {{- range $k, $v := .tags }}
    {{ renderValue $k }} = {{ renderValue $v }}
    {{- end }}
  }
  {{- end }}
}`,
		},
		{
			name: "resource with nested block",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"config": resourceschema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]resourceschema.Attribute{
							"name": resourceschema.StringAttribute{
								Required: true,
							},
							"value": resourceschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  config {
    name = {{ renderValue (required (index .config "name")) }}
    {{- if (index .config "value") }}
    value = {{ renderValue (index .config "value") }}
    {{- end }}
  }
}`,
		},
		{
			name: "resource with list nested block",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"items": resourceschema.ListNestedAttribute{
						Optional: true,
						NestedObject: resourceschema.NestedAttributeObject{
							Attributes: map[string]resourceschema.Attribute{
								"name": resourceschema.StringAttribute{
									Required: true,
								},
								"value": resourceschema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  {{- if .items }}
  items {
    name = {{ renderValue (required (index .items 0 "name")) }}
    {{- if (index .items 0 "value") }}
    value = {{ renderValue (index .items 0 "value") }}
    {{- end }}
  }
  {{- end }}
}`,
		},
		{
			name: "resource with timeouts",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"name": resourceschema.StringAttribute{
						Required: true,
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Update: true,
					}),
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  name = {{ renderValue (required .name) }}
  {{- if .timeouts }}
  timeouts {
    {{- if .timeouts.create }}
    create = {{ renderValue .timeouts.create }}
    {{- end }}
    {{- if .timeouts.update }}
    update = {{ renderValue .timeouts.update }}
    {{- end }}
  }
  {{- end }}
}`,
		},
		{
			name: "resource with depends_on",
			schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"name": resourceschema.StringAttribute{
						Required: true,
					},
					"depends_on": resourceschema.StringAttribute{},
				},
			},
			resourceType: "test_resource",
			kind:         ResourceKindResource,
			want: `resource "test_resource" "{{ required .resource_name }}" {
  name = {{ renderValue (required .name) }}
  {{- if .depends_on }}
  depends_on = [{{- range $i, $dep := .depends_on }}{{if $i}}, {{end}}{{ renderValue $dep }}{{- end }}]
  {{- end }}
}`,
		},
	}

	generator := NewFrameworkTemplateGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := generator.GenerateTemplate(tt.schema, tt.resourceType, tt.kind)
			assert.NoError(t, err)
			assert.Equal(t, normalizeHCL(tt.want), normalizeHCL(got), "Generated template mismatch")
		})
	}
}

func TestFrameworkGenerateTemplateWithInvalidSchema(t *testing.T) {
	generator := NewFrameworkTemplateGenerator()

	// Test with an invalid schema type
	got, err := generator.GenerateTemplate("invalid schema", "test_resource", ResourceKindResource)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema type")
	assert.Equal(t, "", got)
}

func TestFrameworkTimeoutsRendering(t *testing.T) {
	t.Parallel()

	generator := NewFrameworkTemplateGenerator()

	// Create a test field with all timeout types
	timeoutsField := TemplateField{
		Name: "timeouts",
		NestedFields: []TemplateField{
			{Name: "create"},
			{Name: "read"},
			{Name: "update"},
			{Name: "delete"},
		},
	}

	config := generator.extractTimeoutsConfig(timeoutsField)

	// Verify all timeouts are detected
	assert.True(t, config.Create)
	assert.True(t, config.Read)
	assert.True(t, config.Update)
	assert.True(t, config.Delete)

	// Test with only some timeouts
	partialTimeoutsField := TemplateField{
		Name: "timeouts",
		NestedFields: []TemplateField{
			{Name: "read"},
			{Name: "delete"},
		},
	}

	partialConfig := generator.extractTimeoutsConfig(partialTimeoutsField)

	// Verify only specific timeouts are detected
	assert.False(t, partialConfig.Create)
	assert.True(t, partialConfig.Read)
	assert.False(t, partialConfig.Update)
	assert.True(t, partialConfig.Delete)
}
