package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
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

func TestLimitedAvailabilitySchemaDescription(t *testing.T) {
	code, err := genSchema(
		&Definition{
			LimitedAvailability: lo.ToPtr(true),
			Resource:            &SchemaMeta{},
		},
		resourceType,
		&Item{
			Name:        "foo",
			Description: "Does stuff",
			Type:        SchemaTypeObject,
			Properties:  map[string]*Item{},
		},
	)
	require.NoError(t, err)

	file := jen.NewFile("test")
	file.Add(code)

	var b bytes.Buffer
	require.NoError(t, file.Render(&b))
	got := b.String()

	require.Contains(t, got, "limited availability")
	require.Contains(t, got, "contact the [sales team](http://aiven.io/contact)")
}

func TestLimitedAvailabilityViewOptions(t *testing.T) {
	code, err := genNewResource(resourceType, &Definition{LimitedAvailability: lo.ToPtr(true), Resource: &SchemaMeta{}}, &Item{}, false)
	require.NoError(t, err)

	file := jen.NewFile("test")
	file.Add(code)

	var b bytes.Buffer
	require.NoError(t, file.Render(&b))
	got := b.String()

	require.NotContains(t, got, "Limited:")
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

			err := fromSchema(scope, &Operation{}, root, tt.schema, ReadHandler)
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

// TestApplySchemaOverride mirrors the aiven_project_vpc datasource overlay: it
// relaxes a Required API field (`project`) to Optional+Computed and adds a
// brand-new alias field (`vpc_id`) that does not exist in the base API schema.
// The test pins down the contract that the FromSchemaOverride flag is set on
// every touched item, that user-set overrides win over the entity-aware
// defaults, that the original root is never mutated, and that items outside
// the overlay are left alone.
func TestApplySchemaOverride(t *testing.T) {
	def := &Definition{
		typeName: "aiven_project_vpc",
		Operations: Operations{
			{ID: "VpcCreate", Type: OperationCreate},
			{ID: "VpcGet", Type: OperationRead},
			{ID: "VpcDelete", Type: OperationDelete},
		},
		IDAttributeComposed: []string{"project", "project_vpc_id"},
		Datasource: &SchemaMeta{
			SchemaOverride: map[string]*Item{
				"project": {
					OverrideOptional: lo.ToPtr(true),
					OverrideComputed: lo.ToPtr(true),
				},
				"vpc_id": {
					Type:               SchemaTypeString,
					OverrideOptional:   lo.ToPtr(true),
					Description:        "The ID of the VPC in `project/project_vpc_id` format.",
					DeprecationMessage: "Use `project_vpc_id` instead.",
					ConflictsWith:      []string{"project", "project_vpc_id"},
				},
			},
		},
	}

	createPath := def.Operations.AppearsInID("VpcCreate", OperationCreate, PathParameter)
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"project": {
				Name:        "project",
				Type:        SchemaTypeString,
				AppearsIn:   createPath,
				Required:    true,
				IDAttribute: true,
				Description: "The name of the project this resource belongs to.",
			},
			"cloud_name": {
				Name:        "cloud_name",
				Type:        SchemaTypeString,
				Description: "Cloud region.",
			},
		},
	}
	setParents(root, nil)

	out, err := applySchemaOverride(def, def.Datasource, root)
	require.NoError(t, err)
	require.NotSame(t, root, out, "applySchemaOverride must not mutate the original root")

	require.False(t, root.Properties["project"].FromSchemaOverride,
		"original root.project must not be tagged from the overlay")
	require.NotContains(t, root.Properties, "vpc_id",
		"original root must not gain overlay-only attributes")

	project := out.Properties["project"]
	require.NotNil(t, project)
	assert.True(t, project.FromSchemaOverride, "project: FromSchemaOverride must be set after overlay merge")
	assert.True(t, project.Optional, "project: user-set optional must win over the entity-aware default")
	assert.True(t, project.Computed, "project: user-set computed must win over the entity-aware default")
	assert.False(t, project.Required, "project: optional+computed override must clear Required")

	vpcID := out.Properties["vpc_id"]
	require.NotNil(t, vpcID, "vpc_id: overlay-introduced attribute must be present in the copy")
	assert.True(t, vpcID.FromSchemaOverride)
	assert.True(t, vpcID.Optional)
	assert.False(t, vpcID.Required)
	assert.Equal(t, SchemaTypeString, vpcID.Type)
	assert.Equal(t, "Use `project_vpc_id` instead.", vpcID.DeprecationMessage)
	assert.Equal(t, []string{"project", "project_vpc_id"}, vpcID.ConflictsWith)

	cloudName := out.Properties["cloud_name"]
	require.NotNil(t, cloudName)
	assert.False(t, cloudName.FromSchemaOverride, "cloud_name: untouched by the overlay")
}

// TestApplySchemaOverrideNoOp asserts that an empty (or nil) overlay returns
// the input root verbatim without a deep copy.
func TestApplySchemaOverrideNoOp(t *testing.T) {
	def := &Definition{Operations: Operations{{ID: "Get", Type: OperationRead}}}
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{}}

	out, err := applySchemaOverride(def, nil, root)
	require.NoError(t, err)
	require.Same(t, root, out, "nil SchemaMeta must short-circuit without copying")

	out, err = applySchemaOverride(def, &SchemaMeta{}, root)
	require.NoError(t, err)
	require.Same(t, root, out, "empty SchemaOverride must short-circuit without copying")
}

// TestDeepCopyItem guards the yaml round-trip used by applySchemaOverride.
// Item carries runtime state on underscore-prefixed yaml tags so the copy must
// preserve every public-facing and internal field set on the original, and
// the copy's Parent pointers must point at the new tree (not the old one).
// If a new Item field is added in the future without a yaml tag, this test
// should fail and alert the author.
func TestDeepCopyItem(t *testing.T) {
	leaf := &Item{
		Name:        "leaf",
		JSONName:    "leaf",
		Type:        SchemaTypeString,
		Description: "leaf description",
		Required:    true,
		Computed:    true,
		AppearsIn:   ReadHandler | ResponseBody,
	}
	arr := &Item{
		Name:     "items",
		JSONName: "items",
		Type:     SchemaTypeArray,
		Items:    leaf,
	}
	root := &Item{
		Name:               "root",
		JSONName:           "root",
		Type:               SchemaTypeObject,
		Description:        "root description",
		DeprecationMessage: "deprecated",
		OverrideOptional:   lo.ToPtr(true),
		OverrideRequired:   lo.ToPtr(false),
		ConflictsWith:      []string{"a", "b"},
		FromSchemaOverride: true,
		UseStateForUnknown: true,
		MinLength:          3,
		MaxLength:          64,
		Enum:               []any{"x", "y"},
		Properties: map[string]*Item{
			"items":     arr,
			"id":        {Name: "id", JSONName: "id", Type: SchemaTypeString, IDAttribute: true, IDAttributePosition: 0},
			"is_active": {Name: "is_active", JSONName: "is_active", Type: SchemaTypeBoolean, Optional: true},
		},
	}
	setParents(root, nil)

	cp, err := deepCopyItem(root)
	require.NoError(t, err)
	require.NotSame(t, root, cp)

	// Parent is yaml:"-" so it is intentionally not serialized; cmp.Diff with
	// Parent ignored is the right contract.
	ignoreParent := cmpopts.IgnoreFields(Item{}, "Parent")
	if diff := cmp.Diff(root, cp, ignoreParent); diff != "" {
		t.Fatalf("deepCopyItem mismatch (-want +got):\n%s", diff)
	}

	assert.Nil(t, cp.Parent, "root: Parent must be nil in the copy")
	assert.Same(t, cp, cp.Properties["items"].Parent, "Parent must be re-wired into the new tree")
	assert.Same(t, cp.Properties["items"], cp.Properties["items"].Items.Parent)

	// Mutating the copy must not leak back into the original.
	cp.Properties["items"].Items.Description = "mutated"
	assert.Equal(t, "leaf description", leaf.Description)
}
