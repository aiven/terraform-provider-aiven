package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenSchemaAddsSingletonValidatorForComputedObjectBlocks(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"config": {
				Name:     "config",
				Type:     SchemaTypeObject,
				Computed: true,
				Properties: map[string]*Item{
					"retention_ms": {
						Name:     "retention_ms",
						Type:     SchemaTypeString,
						Computed: true,
					},
				},
			},
		},
	}
	root.Properties["config"].Parent = root
	root.Properties["config"].Properties["retention_ms"].Parent = root.Properties["config"]

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, `"config": schema.ListNestedBlock{`)
	require.Contains(t, got, `[]validator.List{listvalidator.SizeAtMost(1)}`)
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
