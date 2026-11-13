package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleVariantScope(t *testing.T) {
	t.Parallel()

	item := &Item{
		Name: "settings",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"shared": {Name: "shared", Type: SchemaTypeString},
			"a":      {Name: "a", Type: SchemaTypeString},
			"b":      {Name: "b", Type: SchemaTypeString},
			"c":      {Name: "c", Type: SchemaTypeString},
		},
		Variants: []*Variant{
			{Attributes: []string{"shared", "a"}},
			{Attributes: []string{"b"}},
			{Attributes: []string{"c"}},
		},
	}
	clone := exampleVariantScope(item)

	assert.Contains(t, clone.Properties, "shared")
	assert.Contains(t, clone.Properties, "a")
	assert.NotContains(t, clone.Properties, "b")
	assert.NotContains(t, clone.Properties, "c")
	// Input unchanged
	require.Contains(t, item.Properties, "b")
	require.Contains(t, item.Properties, "c")
}
