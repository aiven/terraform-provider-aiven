package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleMap(t *testing.T) {
	item := &Item{
		Name: "example",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"map": {
				Type:     SchemaTypeObject,
				Required: true,
				Items: &Item{
					Name: "map_value_string",
					Type: SchemaTypeString,
				},
			},
			"map_nested": {
				Type:     SchemaTypeObject,
				Required: true,
				Items: &Item{
					Name: "object",
					Type: SchemaTypeObject,
					Properties: map[string]*Item{
						"object_field": {
							Name: "map_value_object_field",
							Type: SchemaTypeString,
						},
					},
				},
			},
		},
	}

	expected := strings.TrimSpace(`
resource "aiven_example" "example" {
  map = {
    foo = "foo"
  }
  map_nested = {
    foo = {
      object_field = "foo"
    }
  }
}`)
	result, err := exampleRoot(true, item)
	require.NoError(t, err)
	assert.Equal(t, expected, strings.TrimSpace(result))
}
