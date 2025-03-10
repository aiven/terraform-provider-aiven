package template

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSDKTemplate(t *testing.T) {
	tests := []struct {
		name         string
		resource     *schema.Resource
		resourceType string
		kind         ResourceKind
		want         string
	}{
		{
			name: "resource with boolean fields",
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enable_feature": {
						Type:     schema.TypeBool,
						Required: true,
					},
					"disable_feature": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"nested_settings": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"feature_one": {
									Type:     schema.TypeBool,
									Required: true,
								},
								"feature_two": {
									Type:     schema.TypeBool,
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
  enable_feature = {{ .enable_feature }}
  {{- if ne .disable_feature nil }}
  disable_feature = {{ .disable_feature }}
  {{- end }}
  nested_settings {
    feature_one = {{ (index .nested_settings 0 "feature_one") }}
    {{- if ne (index .nested_settings 0 "feature_two") nil }}
    feature_two = {{ (index .nested_settings 0 "feature_two") }}
    {{- end }}
  }
}`,
		},
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
		{
			name: "resource with nested block",
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"config": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"value": {
									Type:     schema.TypeString,
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
  config {
    name = {{ renderValue (required (index .config 0 "name")) }}
    {{- if (index .config 0 "value") }}
    value = {{ renderValue (index .config 0 "value") }}
    {{- end }}
  }
}`,
		},
		{
			name: "resource with list nested block",
			resource: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"items": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"value": {
									Type:     schema.TypeString,
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
	}

	generator := NewSDKTemplateGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := generator.GenerateTemplate(tt.resource, tt.resourceType, tt.kind)
			assert.Equal(t, normalizeHCL(tt.want), normalizeHCL(got), "Generated template mismatch")
		})
	}
}

func TestResourceKindString(t *testing.T) {
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
			kind: ResourceKind("unknown_kind"),
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
