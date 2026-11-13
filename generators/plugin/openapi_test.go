package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeVariants(t *testing.T) {
	t.Parallel()

	doc := &OpenAPIDoc{}

	tests := []struct {
		name    string
		schema  *OASchema
		wantErr string
		assert  func(t *testing.T, merged *OASchema)
	}{
		{
			name: "oneOf intersection required and variant metadata",
			schema: &OASchema{
				OneOf: []*OASchema{
					{
						Type:     SchemaTypeObject,
						Title:    "Azure",
						Required: []string{"base_path", "account"},
						Properties: map[string]*OASchema{
							"base_path": {Type: SchemaTypeString},
							"account":   {Type: SchemaTypeString},
							"z":         {Type: SchemaTypeString},
						},
					},
					{
						Type:     SchemaTypeObject,
						Required: []string{"base_path", "bucket"},
						Properties: map[string]*OASchema{
							"bucket":    {Type: SchemaTypeString},
							"base_path": {Type: SchemaTypeString},
						},
					},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.Equal(t, []string{"base_path"}, merged.Required)
				require.Len(t, merged.variants, 2)
				assert.Equal(t, "Azure", merged.variants[0].title)
				assert.Equal(t, "Variant 2", merged.variants[1].title)
				assert.Equal(t, []string{"account", "base_path", "z"}, merged.variants[0].properties)
				assert.ElementsMatch(t, []string{"account"}, merged.variants[0].required)
				assert.False(t, merged.anyVariant)
			},
		},
		{
			name: "anyOf sets anyVariant",
			schema: &OASchema{
				AnyOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.True(t, merged.anyVariant)
			},
		},
		{
			name: "allOf unions required without variants",
			schema: &OASchema{
				AllOf: []*OASchema{
					{Type: SchemaTypeObject, Required: []string{"a"}, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
					{Type: SchemaTypeObject, Required: []string{"b"}, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.ElementsMatch(t, []string{"a", "b"}, merged.Required)
				assert.Empty(t, merged.variants)
			},
		},
		{
			name: "integer and number conflict",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"n": {Type: SchemaTypeInteger}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"n": {Type: SchemaTypeNumber}}},
				},
			},
			wantErr: `n: the variants have different type: "integer" and "number"`,
		},
		{
			name: "bounds mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {
						Type: SchemaTypeString, Minimum: json.Number("2"), Maximum: json.Number("10"), MinLength: 1, MaxLength: 5,
					}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {
						Type: SchemaTypeString, Minimum: json.Number("1"), Maximum: json.Number("20"), MinLength: 3, MaxLength: 0,
					}}},
				},
			},
			wantErr: `s: the variants have different minLength: 1 and 3`,
		},
		{
			name: "equal bounds kept",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {
						Type: SchemaTypeString, Minimum: json.Number("1"), Maximum: json.Number("10"), MinLength: 2, MaxLength: 5,
					}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {
						Type: SchemaTypeString, Minimum: json.Number("1"), Maximum: json.Number("10"), MinLength: 2, MaxLength: 5,
					}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				p := merged.Properties["s"]
				assert.Equal(t, json.Number("1"), p.Minimum)
				assert.Equal(t, json.Number("10"), p.Maximum)
				assert.Equal(t, 2, p.MinLength)
				assert.Equal(t, 5, p.MaxLength)
			},
		},
		{
			name: "missing bound conflicts",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeInteger, Minimum: json.Number("1")}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeInteger}}},
				},
			},
			wantErr: `s: the variants have different minimum: 1 and 0`,
		},
		{
			name: "enum mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString, Enum: []any{"a", "b"}}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString, Enum: []any{"b", "c"}}}},
				},
			},
			wantErr: `e: the variants have different enum: ["a","b"] and ["b","c"]`,
		},
		{
			name: "enum missing in one variant",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString, Enum: []any{"a"}}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString}}},
				},
			},
			wantErr: `e: the variants have different enum: ["a"] and null`,
		},
		{
			name: "enum same values different order",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString, Enum: []any{"a", "b"}}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"e": {Type: SchemaTypeString, Enum: []any{"b", "a"}}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.ElementsMatch(t, []any{"a", "b"}, merged.Properties["e"].Enum)
			},
		},
		{
			name: "pattern equal when variants agree",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeString, Pattern: "^a$"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeString, Pattern: "^a$"}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.Equal(t, "^a$", merged.Properties["s"].Pattern)
			},
		},
		{
			name: "pattern mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeString, Pattern: "^a$"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"s": {Type: SchemaTypeString, Pattern: "^b$"}}},
				},
			},
			wantErr: `s: the variants have different pattern: "^a$" and "^b$"`,
		},
		{
			name: "type conflict",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeInteger}}},
				},
			},
			wantErr: `x: the variants have different type: "string" and "integer"`,
		},
		{
			name: "description missing in one variant",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Description: "from first"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				p := merged.Properties["x"]
				assert.Equal(t, "from first", p.Description)
				assert.Empty(t, p.descriptionConflict)
			},
		},
		{
			name: "description mismatch sets conflict",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Description: "a"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Description: "b"}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				p := merged.Properties["x"]
				assert.Equal(t, "a", p.Description)
				assert.Equal(t, `the variants have different description: "a" and "b"`, p.descriptionConflict)
			},
		},
		{
			name: "title mismatch sets conflict",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Title: "one"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Title: "two"}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				p := merged.Properties["x"]
				assert.Equal(t, "one", p.Title)
				assert.Equal(t, `the variants have different title: "one" and "two"`, p.descriptionConflict)
			},
		},
		{
			name: "example mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Example: "a"}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Example: "b"}}},
				},
			},
			wantErr: `x: the variants have different example: "a" and "b"`,
		},
		{
			name: "sensitive flag mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Sensitive: true}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Sensitive: false}}},
				},
			},
			wantErr: `x: the variants have different _secure: true and false`,
		},
		{
			name: "items in one variant only",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeArray}}},
				},
			},
			wantErr: `x: the variants have different items: true and false`,
		},
		{
			name: "items type mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeInteger}}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}}}},
				},
			},
			wantErr: `x: items: the variants have different type: "integer" and "string"`,
		},
		{
			name: "scalar variant unchanged",
			schema: &OASchema{OneOf: []*OASchema{
				{Type: SchemaTypeBoolean},
				{Type: SchemaTypeObject, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
			}},
			assert: func(t *testing.T, merged *OASchema) {
				require.Len(t, merged.OneOf, 2)
				assert.Empty(t, merged.Properties)
			},
		},
		{
			name: "additionalProperties closed only when all closed",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, AdditionalProperties: &OAAdditionalProperties{}, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
					{Type: SchemaTypeObject, AdditionalProperties: &OAAdditionalProperties{Allowed: true}, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				assert.Nil(t, merged.AdditionalProperties)
			},
		},
		{
			name: "all variants closed",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, AdditionalProperties: &OAAdditionalProperties{}, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
					{Type: SchemaTypeObject, AdditionalProperties: &OAAdditionalProperties{}, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				require.NotNil(t, merged.AdditionalProperties)
				assert.False(t, merged.AdditionalProperties.Allowed)
			},
		},
		{
			name: "nested required mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Required: []string{"a"}, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}, "shared": {Type: SchemaTypeString}},
					}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Required: []string{"b"}, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}, "shared": {Type: SchemaTypeString}},
					}}},
				},
			},
			wantErr: `nested: the variants have different required: ["a"] and ["b"]`,
		},
		{
			name: "nested property keys mismatch",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}},
					}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}, "b": {Type: SchemaTypeString}},
					}}},
				},
			},
			wantErr: `nested: the variants have different properties: ["a"] and ["a","b"]`,
		},
		{
			name: "nested required order ignored",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Required: []string{"a", "b"}, Properties: map[string]*OASchema{
							"a": {Type: SchemaTypeString}, "b": {Type: SchemaTypeString},
						},
					}}},
					{Type: SchemaTypeObject, Properties: map[string]*OASchema{"nested": {
						Type: SchemaTypeObject, Required: []string{"b", "a"}, Properties: map[string]*OASchema{
							"a": {Type: SchemaTypeString}, "b": {Type: SchemaTypeString},
						},
					}}},
				},
			},
			assert: func(t *testing.T, merged *OASchema) {
				nested := merged.Properties["nested"]
				require.NotNil(t, nested)
				assert.ElementsMatch(t, []string{"a", "b"}, nested.Required)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			merged, err := doc.mergeVariants(tt.schema)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, merged)
			}
		})
	}
}

func TestMergeVariantsSharedPropertyRef(t *testing.T) {
	t.Parallel()

	doc := &OpenAPIDoc{}
	doc.Components.Schemas = map[string]*OASchema{
		"SharedString": {Type: SchemaTypeString, Pattern: "^x$"},
	}
	schema := &OASchema{
		OneOf: []*OASchema{
			{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Ref: "#/components/schemas/SharedString"}}},
			{Type: SchemaTypeObject, Properties: map[string]*OASchema{"x": {Type: SchemaTypeString, Pattern: "^x$"}}},
		},
	}
	merged, err := doc.mergeVariants(schema)
	require.NoError(t, err)
	assert.Equal(t, "^x$", merged.Properties["x"].Pattern)
}

func TestResolveSchemaMergesComponentOneOf(t *testing.T) {
	t.Parallel()

	doc := &OpenAPIDoc{}
	comp := &OASchema{
		OneOf: []*OASchema{
			{Type: SchemaTypeObject, Properties: map[string]*OASchema{"a": {Type: SchemaTypeString}}},
			{Type: SchemaTypeObject, Properties: map[string]*OASchema{"b": {Type: SchemaTypeString}}},
		},
	}
	doc.Components.Schemas = map[string]*OASchema{"M": comp}

	resolved, err := doc.resolveSchema(&OASchema{Ref: "#/components/schemas/M"})
	require.NoError(t, err)
	assert.Contains(t, resolved.Properties, "a")
	assert.Len(t, comp.OneOf, 2)
}
