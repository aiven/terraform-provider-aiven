package template

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFrameworkGenerateTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		schema       interface{}
		resourceType string
		kind         ResourceKind
		want         string
	}{
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
  name = {{ renderValue (required .name) }}
  {{- if .description }}
  description = {{ renderValue .description }}
  {{- end }}
}`,
		},
		{
			name: "data source with required fields",
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
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
    name = {{ renderValue (required .name) }}
    {{- if .value }}
    value = {{ renderValue .value }}
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
    name = {{ renderValue (required .name) }}
    {{- if .value }}
    value = {{ renderValue .value }}
    {{- end }}
  }
  {{- end }}
}`,
		},
	}

	generator := NewFrameworkSchemaTemplateGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := generator.GenerateTemplate(tt.schema, tt.resourceType, tt.kind)
			assert.Equal(t, normalizeHCL(tt.want), normalizeHCL(got), "Generated template mismatch")
		})
	}
}

func TestFrameworkGenerateTemplateWithInvalidSchema(t *testing.T) {
	t.Parallel()

	generator := NewFrameworkSchemaTemplateGenerator()

	// Test with an invalid schema type
	got := generator.GenerateTemplate("invalid schema", "test_resource", ResourceKindResource)
	want := `resource "test_resource" "{{ required .resource_name }}" {
  # Error: Framework generator received non-Framework schema
}`

	assert.Equal(t, normalizeHCL(want), normalizeHCL(got), "Generated template for invalid schema mismatch")
}
