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
		typeName:       "aiven_example",
		Resource:       &SchemaMeta{},
		IdentityFields: []string{"project", "example_id"},
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

func TestDocImportIdentity(t *testing.T) {
	def := &Definition{
		typeName:               "aiven_example",
		Resource:               &SchemaMeta{},
		IdentityFields:         []string{"project", "example_id", "region"},
		IdentityFieldsOptional: []string{"region"},
	}
	root := &Item{Name: "example", Properties: map[string]*Item{
		"project":    {Name: "project", IDAttribute: true},
		"example_id": {Name: "example_id", IDAttribute: true, IDAttributePosition: 1},
		"region":     {Name: "region", IDAttribute: true, IDAttributePosition: 2},
	}}
	got, ok := docImport(def, resourceType, root)
	got = strings.Join(strings.Fields(got), " ")
	require.True(t, ok)
	require.Contains(t, got, "terraform import aiven_example.example PROJECT/EXAMPLE_ID")
	require.Contains(t, got, "Terraform 1.12")
	require.Contains(t, got, "identity = {")
	require.Contains(t, got, `project = "PROJECT"`)
	require.Contains(t, got, `example_id = "EXAMPLE_ID"`)
	require.Contains(t, got, `region = "REGION" # Optional`)
	_, ok = docImport(def, datasourceType, root)
	require.False(t, ok)
}
