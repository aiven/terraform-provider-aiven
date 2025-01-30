package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operations func(*testing.T, *registry)
		validate   func(*testing.T, *registry)
	}{
		{
			name: "add and render simple template",
			operations: func(t *testing.T, r *registry) {
				err := r.addTemplate("test", `Hello {{ .name }}`)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, r *registry) {
				result, err := r.render("test", map[string]any{"name": "World"})
				require.NoError(t, err)
				assert.Equal(t, "Hello World", result)
			},
		},
		{
			name: "add template with required field",
			operations: func(t *testing.T, r *registry) {
				err := r.addTemplate("test", `{{ required .name }}`)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, r *registry) {
				_, err := r.render("test", map[string]any{})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required field is missing")
			},
		},
		{
			name: "render non-existent template",
			validate: func(t *testing.T, r *registry) {
				_, err := r.render("non-existent", nil)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "does not exist for resource")
			},
		},
		{
			name: "invalid template syntax",
			operations: func(t *testing.T, r *registry) {
				err := r.addTemplate("invalid", `{{ .name `)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse template")
			},
		},
		{
			name: "add custom function",
			operations: func(t *testing.T, r *registry) {
				r.addFunction("double", func(x int) int { return x * 2 })
				err := r.addTemplate("test", `{{ double .number }}`)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, r *registry) {
				result, err := r.render("test", map[string]any{"number": 5})
				require.NoError(t, err)
				assert.Equal(t, "10", result)
			},
		},
		{
			name: "render with Value type",
			operations: func(t *testing.T, r *registry) {
				err := r.addTemplate("test", `{{ renderValue .val }}`)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, r *registry) {
				literal := Value{Value: "test", IsLiteral: true}
				result, err := r.render("test", map[string]any{"val": literal})
				require.NoError(t, err)
				assert.Equal(t, `"test"`, result)

				reference := Value{Value: "test.reference", IsLiteral: false}
				result, err = r.render("test", map[string]any{"val": reference})
				require.NoError(t, err)
				assert.Equal(t, "test.reference", result)
			},
		},
		{
			name: "render with different value types",
			operations: func(t *testing.T, r *registry) {
				err := r.addTemplate("test", `{{ renderValue .val }}`)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, r *registry) {
				cases := []struct {
					name     string
					input    interface{}
					expected string
				}{
					{"string", "hello", `"hello"`},
					{"int", 42, "42"},
					{"float", 3.14, "3.14"},
					{"bool", true, "true"},
				}

				for _, tc := range cases {
					t.Run(tc.name, func(t *testing.T) {
						result, err := r.render("test", map[string]any{"val": tc.input})
						require.NoError(t, err)
						assert.Equal(t, tc.expected, result)
					})
				}
			},
		},
		{
			name: "get available templates",
			operations: func(_ *testing.T, r *registry) {
				r.mustAddTemplate("b", "template B")
				r.mustAddTemplate("a", "template A")
				r.mustAddTemplate("c", "template C")
			},
			validate: func(t *testing.T, r *registry) {
				templates := r.getAvailableTemplates()
				assert.Equal(t, []string{"a", "b", "c"}, templates)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := newTemplateRegistry(t)

			if tt.operations != nil {
				tt.operations(t, reg)
			}

			if tt.validate != nil {
				tt.validate(t, reg)
			}
		})
	}
}

func TestRegistryTemplateInterpolation(t *testing.T) {
	t.Parallel()

	reg := newTemplateRegistry(t)

	// Test complex template with multiple interpolations
	tmpl := `
resource "test" "{{ .name }}" {
	field1 = {{ renderValue .field1 }}
	{{- if .optional_field }}
	field2 = {{ renderValue .optional_field }}
	{{- end }}
	nested {
		{{- range .items }}
		item = {{ renderValue . }}
		{{- end }}
	}
}`

	err := reg.addTemplate("complex", tmpl)
	require.NoError(t, err)

	config := map[string]any{
		"name":           "test_resource",
		"field1":         "required_value",
		"optional_field": "optional_value",
		"items":          []string{"item1", "item2", "item3"},
	}

	result, err := reg.render("complex", config)
	require.NoError(t, err)

	// Verify template rendering
	assert.Contains(t, result, `resource "test" "test_resource"`)
	assert.Contains(t, result, `field1 = "required_value"`)
	assert.Contains(t, result, `field2 = "optional_value"`)
	assert.Contains(t, result, `item = "item1"`)
	assert.Contains(t, result, `item = "item2"`)
	assert.Contains(t, result, `item = "item3"`)
}
