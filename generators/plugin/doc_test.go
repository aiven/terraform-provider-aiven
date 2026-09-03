package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImportCommandsUsesComposedIDByDefault(t *testing.T) {
	t.Parallel()

	def := &Definition{
		typeName:            "aiven_example",
		Resource:            &SchemaMeta{},
		IDAttributeComposed: []string{"project", "example_id"},
	}
	root := &Item{Name: "example"}

	formats := importIDFormats(def, root)
	require.Equal(t, []string{"PROJECT/EXAMPLE_ID"}, formats)
	require.Equal(t,
		"terraform import aiven_example.example PROJECT/EXAMPLE_ID\n",
		importCommands(def, formats),
	)
}

func TestDocImportRendersConfiguredIDFormats(t *testing.T) {
	t.Parallel()

	def := &Definition{
		typeName: "aiven_example",
		Resource: &SchemaMeta{ImportIDFormats: []string{
			"PROJECT/EXAMPLE_ID",
			"PROJECT/EXAMPLE_ID/REGION",
		}},
	}

	got, ok := docImport(def, resourceType, &Item{Name: "example"})
	require.True(t, ok)
	require.Equal(t,
		"## Import\n\n"+
			"Import is supported using one of the following formats:\n\n"+
			"```shell\n"+
			"terraform import aiven_example.example PROJECT/EXAMPLE_ID\n"+
			"terraform import aiven_example.example PROJECT/EXAMPLE_ID/REGION\n"+
			"```",
		got,
	)
}

func TestSchemaMetaImportIDFormatsYAML(t *testing.T) {
	t.Parallel()

	var def Definition
	decoder := yaml.NewDecoder(strings.NewReader(`resource:
  importIDFormats:
    - PROJECT/EXAMPLE_ID
    - PROJECT/EXAMPLE_ID/REGION
`))
	decoder.KnownFields(true)
	require.NoError(t, decoder.Decode(&def))
	require.Equal(t,
		[]string{"PROJECT/EXAMPLE_ID", "PROJECT/EXAMPLE_ID/REGION"},
		def.Resource.ImportIDFormats,
	)
}
