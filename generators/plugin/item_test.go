package main

import (
	"fmt"
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

			err := fromSchema(scope, Operation{}, root, tt.schema, ReadHandler)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, root.Properties, transformer); diff != "" {
				t.Errorf("properties mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOperationsIDAppearsIn(t *testing.T) {
	operations := Operations{
		{ID: "FooCreate", Type: OperationCreate},
		{ID: "FooRead", Type: OperationRead},
		{ID: "FooUpdate", Type: OperationUpdate},
		{ID: "FooDelete", Type: OperationDelete},
	}

	sources := listSources()
	handlers := operationToHandler()
	seen := make(map[AppearsIn]bool)

	for _, op := range operations {
		for i, source := range sources {
			t.Run(fmt.Sprintf("%s_%d_source", op.ID, i), func(t *testing.T) {
				this := operations.AppearsInID(op.ID, op.Type, source)
				require.NotEqual(t, 0, this)
				require.Positive(t, this&handlers[op.Type])
				require.Positive(t, this&source)

				// Same as by handler type
				require.Equal(t,
					fmt.Sprintf("%b", this),
					fmt.Sprintf("%b", operations.AppearsInHandler(op.Type, source)),
				)

				// Proves the mask doesn't contain other operations: create, update, etc
				for opType, a := range handlers {
					if opType != op.Type {
						require.EqualValues(t, 0, this&a)
					}
				}

				// Proves the mask doesn't contain other sources: RequestBody, ResponseBody, etc
				for _, s := range sources {
					if s != source {
						require.EqualValues(t, 0, this&s)
					}
				}

				// Stores to make sure all values are unique
				require.False(t, seen[this])
				seen[this] = true
			})
		}
	}
}

// TestOperationsIDAppearsInUpsert tests when the same operation ID is used for both create and update (upsert pattern)
func TestOperationsIDAppearsInUpsert(t *testing.T) {
	operations := Operations{
		{ID: "FooUpsert", Type: OperationCreate},
		{ID: "FooRead", Type: OperationRead},
		{ID: "FooUpsert", Type: OperationUpdate}, // Same ID as create
		{ID: "FooDelete", Type: OperationDelete},
	}

	sources := listSources()
	handlers := operationToHandler()
	seen := make(map[AppearsIn]bool)

	for _, op := range operations {
		for i, source := range sources {
			t.Run(fmt.Sprintf("%s_%s_%d_source", op.ID, op.Type, i), func(t *testing.T) {
				this := operations.AppearsInID(op.ID, op.Type, source)
				require.NotEqual(t, 0, this)
				require.Positive(t, this&handlers[op.Type])
				require.Positive(t, this&source)

				// Proves the mask doesn't contain other operations: create, update, etc
				for opType, a := range handlers {
					if opType != op.Type {
						require.EqualValues(t, 0, this&a)
					}
				}

				// Proves the mask doesn't contain other sources: RequestBody, ResponseBody, etc
				for _, s := range sources {
					if s != source {
						require.EqualValues(t, 0, this&s)
					}
				}

				// Stores to make sure all values are unique
				require.False(t, seen[this])
				seen[this] = true
			})
		}
	}

	// Test that AppearsInHandler merges both create and update when same ID is used
	for _, source := range sources {
		createMask := operations.AppearsInID("FooUpsert", OperationCreate, source)
		updateMask := operations.AppearsInID("FooUpsert", OperationUpdate, source)

		// Create and update masks should be different
		require.NotEqual(t, createMask, updateMask)

		// AppearsInHandler should return only the specific handler's mask
		require.Equal(t,
			fmt.Sprintf("%b", createMask),
			fmt.Sprintf("%b", operations.AppearsInHandler(OperationCreate, source)),
		)
		require.Equal(t,
			fmt.Sprintf("%b", updateMask),
			fmt.Sprintf("%b", operations.AppearsInHandler(OperationUpdate, source)),
		)
	}
}
