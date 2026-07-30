package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// rootWithObject builds a resource root that holds a single object property.
func rootWithObject(object *Item) *Item {
	root := &Item{
		Name:       "root",
		Type:       SchemaTypeObject,
		Properties: map[string]*Item{object.Name: object},
	}

	object.Parent = root
	for _, v := range object.Properties {
		v.Parent = object
	}
	return root
}

func genResourceSchemaCode(t *testing.T, root *Item) string {
	t.Helper()

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)
	return renderCode(t, code)
}

// A read-only collection renders as a computed attribute, see Item.RendersAsAttribute.
func TestGenSchemaRendersReadOnlyObjectAsAttribute(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "current_deployment",
		Type:     SchemaTypeObject,
		Computed: true,
		Properties: map[string]*Item{
			"status": {Name: "status", Type: SchemaTypeString, Computed: true},
		},
	}))

	require.Contains(t, got, `"current_deployment": schema.ListNestedAttribute{`)
	// Validators only see the configuration.
	require.NotContains(t, got, "listvalidator")
}

// A configurable object stays a block, and keeps the singleton validator.
func TestGenSchemaAddsSingletonValidatorForOptionalObjectBlocks(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "config",
		Type:     SchemaTypeObject,
		Optional: true,
		Computed: true,
		Properties: map[string]*Item{
			"retention_ms": {Name: "retention_ms", Type: SchemaTypeString, Optional: true, Computed: true},
		},
	}))

	require.Contains(t, got, `"config": schema.ListNestedBlock{`)
	require.Contains(t, got, `[]validator.List{listvalidator.SizeAtMost(1)}`)
}

// One settable field is enough to keep the block syntax.
func TestGenSchemaKeepsObjectWithOptionalFieldAsBlock(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "config",
		Type:     SchemaTypeObject,
		Computed: true,
		Properties: map[string]*Item{
			"retention_ms": {Name: "retention_ms", Type: SchemaTypeString, Optional: true},
		},
	}))

	require.Contains(t, got, `"config": schema.ListNestedBlock{`)
}

func TestGenSchemaAddsCIDRValidatorForCIDRFormat(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"network_cidr": {
				Name:     "network_cidr",
				Type:     SchemaTypeString,
				Required: true,
				Format:   "cidr",
			},
		},
	}
	root.Properties["network_cidr"].Parent = root

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, `"network_cidr": schema.StringAttribute{`)
	require.Contains(t, got, `[]validator.String{validators.CIDR()}`)
}
