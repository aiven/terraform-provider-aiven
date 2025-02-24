package template

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resource     *schema.Resource
		resourceType string
		kind         ResourceKind
		want         string
	}{
		{
			name: "basic resource with required string field",
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
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
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"description": {
						Type:     schema.TypeString,
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
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:     schema.TypeString,
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
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"tags": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
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
			name: "resource with timeouts",
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
				Timeouts: &schema.ResourceTimeout{
					Create: schema.DefaultTimeout(30 * time.Second),
					Update: schema.DefaultTimeout(30 * time.Second),
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
	}

	generator := NewSchemaTemplateGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := generator.GenerateTemplate(tt.resource, tt.resourceType, tt.kind)
			assert.Equal(t, normalizeHCL(tt.want), normalizeHCL(got), "Generated template mismatch")
		})
	}
}

func TestResourceKindString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind ResourceKind
		want string
	}{
		{
			name: "resource kind",
			kind: ResourceKindResource,
			want: "resource",
		},
		{
			name: "data source kind",
			kind: ResourceKindDataSource,
			want: "data",
		},
		{
			name: "unknown kind",
			kind: ResourceKind(999),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.kind.String()
			assert.Equal(t, tt.want, got, "ResourceKind string representation mismatch")
		})
	}
}
