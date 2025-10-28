package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemPath(t *testing.T) {
	items := &Item{
		Name: "array",
		Type: SchemaTypeString,
	}
	array := &Item{
		Name:  "array",
		Type:  SchemaTypeArray,
		Items: items,
	}
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"array": array,
		},
	}

	array.Parent = root
	items.Parent = array
	assert.Equal(t, "array", array.Path())
	assert.Equal(t, "array", items.Path())
}

// TestItemRemoveElements verifies that elements are correctly removed from nested structures.
func TestItemRemoveElements(t *testing.T) {
	tests := []struct {
		name   string
		schema *OASchema
		remove []string
		want   map[string]*Item
	}{
		{
			name: "array of objects",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"plans": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"count": {Type: SchemaTypeInteger},
								"name":  {Type: SchemaTypeString},
							},
						},
					},
				},
			},
			remove: []string{"plans/count"},
			want: map[string]*Item{
				"plans": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"name": {Properties: map[string]*Item{}}, // "count" removed
						},
					},
				},
			},
		},
		{
			name: "nested arrays",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"backups": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"files": {
									Type: SchemaTypeArray,
									Items: &OASchema{
										Type: SchemaTypeObject,
										Properties: map[string]*OASchema{
											"size": {Type: SchemaTypeInteger},
											"name": {Type: SchemaTypeString},
										},
									},
								},
							},
						},
					},
				},
			},
			remove: []string{"backups/files/size"},
			want: map[string]*Item{
				"backups": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"files": {
								Properties: map[string]*Item{},
								Items: &Item{
									Properties: map[string]*Item{
										"name": {Properties: map[string]*Item{}}, // "size" removed
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "map of objects",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"regions": {
						Type: SchemaTypeObject,
						AdditionalProperties: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"zone":   {Type: SchemaTypeString},
								"active": {Type: SchemaTypeBoolean},
							},
						},
					},
				},
			},
			remove: []string{"regions/zone"},
			want: map[string]*Item{
				"regions": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"active": {Properties: map[string]*Item{}}, // "zone" removed
						},
					},
				},
			},
		},
	}

	transformer := cmp.Transformer("Item", func(i *Item) any {
		if i == nil {
			return nil
		}

		return struct {
			Properties map[string]*Item
			Items      *Item
		}{
			Properties: i.Properties,
			Items:      i.Items,
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Item{Name: "root", Type: SchemaTypeObject, Properties: make(map[string]*Item)}
			scope := &Scope{Definition: &Definition{Remove: tt.remove}}

			err := fromSchema(scope, root, tt.schema, ReadHandler)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, root.Properties, transformer); diff != "" {
				t.Errorf("properties mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
