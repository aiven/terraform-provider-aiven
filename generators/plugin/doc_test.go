package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDocImport(t *testing.T) {
	t.Run("composed ID", func(t *testing.T) {
		def := &Definition{
			typeName: "aiven_example", Resource: &SchemaMeta{},
			IDAttributeComposed: []string{"project", "example_id"},
		}
		got, ok := docImport(def, resourceType, &Item{Name: "example"})
		require.True(t, ok)
		require.Equal(t, "## Import\n\nImport is supported using the following syntax:\n\n```shell\n"+
			"terraform import aiven_example.example PROJECT/EXAMPLE_ID\n```", got)
	})

	t.Run("custom formats", func(t *testing.T) {
		var def Definition
		decoder := yaml.NewDecoder(strings.NewReader(`resource:
  importIDFormats:
    - PROJECT/EXAMPLE_ID
    - PROJECT/EXAMPLE_ID/REGION
`))
		decoder.KnownFields(true)
		require.NoError(t, decoder.Decode(&def))
		def.typeName = "aiven_example"
		root := &Item{Name: "example"}
		commands := "terraform import aiven_example.example PROJECT/EXAMPLE_ID\n" +
			"# Alternatively:\n# terraform import aiven_example.example PROJECT/EXAMPLE_ID/REGION\n"
		require.Equal(t, commands, importCommands(&def, importIDFormats(&def, root)))
		got, ok := docImport(&def, resourceType, root)
		require.True(t, ok)
		require.Equal(t, "## Import\n\nImport is supported using one of the following formats:\n\n```shell\n"+
			commands+"```", got)
	})
}
