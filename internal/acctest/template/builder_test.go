package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompositionBuilder_Remove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialSetup func(*CompositionBuilder)
		resourcePath string
		expectedSize int
		validateFunc func(*testing.T, *CompositionBuilder)
	}{
		{
			name: "remove existing resource",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "test-project",
				})
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test2",
					"project":       "test-project",
				})
			},
			resourcePath: "aiven_kafka.test1",
			expectedSize: 1,
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
				assert.Equal(t, "test2", b.compositions[0].Config["resource_name"])
			},
		},
		{
			name: "remove non-existing resource",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "test-project",
				})
			},
			resourcePath: "aiven_kafka.non_existing",
			expectedSize: 1,
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
				assert.Equal(t, "test1", b.compositions[0].Config["resource_name"])
			},
		},
		{
			name: "remove with invalid resource path",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "test-project",
				})
			},
			resourcePath: "invalid_path",
			expectedSize: 1,
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
			},
		},
		{
			name:         "remove from empty builder",
			initialSetup: func(_ *CompositionBuilder) {},
			resourcePath: "aiven_kafka.test",
			expectedSize: 0,
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Empty(t, b.compositions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := newTemplateRegistry(t)
			reg.mustAddTemplate("resource.aiven_kafka", resourceTemplate)

			builder := &CompositionBuilder{
				registry:     reg,
				compositions: make([]compositionEntry, 0),
			}

			tt.initialSetup(builder)
			builder.Remove(tt.resourcePath)

			assert.Len(t, builder.compositions, tt.expectedSize)
			if tt.validateFunc != nil {
				tt.validateFunc(t, builder)
			}
		})
	}
}

func TestCompositionBuilder_Replace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialSetup func(*CompositionBuilder)
		resourcePath string
		newConfig    map[string]any
		validateFunc func(*testing.T, *CompositionBuilder)
	}{
		{
			name: "replace existing resource",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "old-project",
				})
			},
			resourcePath: "aiven_kafka.test1",
			newConfig: map[string]any{
				"project": "new-project",
				"plan":    "business-4",
			},
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
				assert.Equal(t, "test1", b.compositions[0].Config["resource_name"])
				assert.Equal(t, "new-project", b.compositions[0].Config["project"])
				assert.Equal(t, "business-4", b.compositions[0].Config["plan"])
			},
		},
		{
			name: "replace non-existing resource",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "test-project",
				})
			},
			resourcePath: "aiven_kafka.test2",
			newConfig: map[string]any{
				"project": "new-project",
			},
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 2)
				// Original resource unchanged
				assert.Equal(t, "test1", b.compositions[0].Config["resource_name"])
				assert.Equal(t, "test-project", b.compositions[0].Config["project"])
				// New resource added
				assert.Equal(t, "test2", b.compositions[1].Config["resource_name"])
				assert.Equal(t, "new-project", b.compositions[1].Config["project"])
			},
		},
		{
			name: "replace with invalid resource path",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "test-project",
				})
			},
			resourcePath: "invalid_path",
			newConfig: map[string]any{
				"project": "new-project",
			},
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
				assert.Equal(t, "test1", b.compositions[0].Config["resource_name"])
				assert.Equal(t, "test-project", b.compositions[0].Config["project"])
			},
		},
		{
			name: "replace with nested configuration",
			initialSetup: func(b *CompositionBuilder) {
				b.AddResource("aiven_kafka", map[string]any{
					"resource_name": "test1",
					"project":       "old-project",
					"user_config": []map[string]any{{
						"version": "2.0",
					}},
				})
			},
			resourcePath: "aiven_kafka.test1",
			newConfig: map[string]any{
				"project": "new-project",
				"user_config": []map[string]any{{
					"version": "3.0",
					"params": map[string]any{
						"key": "value",
					},
				}},
			},
			validateFunc: func(t *testing.T, b *CompositionBuilder) {
				assert.Len(t, b.compositions, 1)
				assert.Equal(t, "test1", b.compositions[0].Config["resource_name"])
				assert.Equal(t, "new-project", b.compositions[0].Config["project"])

				userConfig := b.compositions[0].Config["user_config"].([]map[string]any)
				assert.Equal(t, "3.0", userConfig[0]["version"])
				assert.Equal(t, "value", userConfig[0]["params"].(map[string]any)["key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := newTemplateRegistry(t)
			reg.mustAddTemplate("resource.aiven_kafka", resourceTemplate)

			builder := &CompositionBuilder{
				registry:     reg,
				compositions: make([]compositionEntry, 0),
			}

			tt.initialSetup(builder)
			builder.Replace(tt.resourcePath, tt.newConfig)

			if tt.validateFunc != nil {
				tt.validateFunc(t, builder)
			}
		})
	}
}

const resourceTemplate = `resource "aiven_kafka" "{{ .resource_name }}" {
	{{- if .project }}
	project = {{ renderValue .project }}
	{{- end }}
	{{- if .plan }}
	plan = {{ renderValue .plan }}
	{{- end }}
	{{- if .user_config }}
	user_config {
		{{- range $idx, $conf := .user_config }}
		{{- if .version }}
		version = {{ renderValue .version }}
		{{- end }}
		{{- if .params }}
		params = {{ renderValue .params }}
		{{- end }}
		{{- end }}
	}
	{{- end }}
}`
