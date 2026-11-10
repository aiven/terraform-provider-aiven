package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocImport(t *testing.T) {
	t.Run("composed ID", func(t *testing.T) {
		def := &Definition{
			typeName: "aiven_example", Resource: &SchemaMeta{},
			IdentityFields: []string{"project", "example_id"},
		}
		got, ok := docImport(def, resourceType, &Item{Name: "example"})
		require.True(t, ok)
		require.Contains(t, got, "## Import\n\nImport is supported using the following syntax:\n\n```shell\n"+
			"terraform import aiven_example.example PROJECT/EXAMPLE_ID\n```")
	})

	t.Run("optional identity suffix", func(t *testing.T) {
		def := &Definition{
			typeName: "aiven_example", Resource: &SchemaMeta{},
			IdentityFields:         []string{"project", "example_id", "region", "zone"},
			IdentityFieldsOptional: []string{"region", "zone"},
		}
		root := &Item{Name: "example"}
		require.Equal(t, "terraform import aiven_example.example PROJECT/EXAMPLE_ID/REGION/ZONE\n", importCommand(def, root))
		got, ok := docImport(def, resourceType, root)
		require.True(t, ok)
		require.Contains(t, got, "## Import\n\nImport is supported using the following syntax:\n\n```shell\n"+
			"terraform import aiven_example.example PROJECT/EXAMPLE_ID/REGION/ZONE\n```")
	})

	t.Run("custom example preserves resource identity", func(t *testing.T) {
		const example = `terraform import aiven_example.example PROJECT/SERVICE_NAME
# For multiple connections, run this instead:
# terraform import aiven_example.example PROJECT/SERVICE_NAME/CONNECTION_ID`
		def := &Definition{
			typeName:       "aiven_example",
			Resource:       &SchemaMeta{ImportExample: example},
			IdentityFields: []string{"project", "service_name"},
		}
		root := &Item{Name: "example", Properties: map[string]*Item{
			"project":       {Name: "project", IDAttribute: true},
			"service_name":  {Name: "service_name", IDAttribute: true, IDAttributePosition: 1},
			"connection_id": {Name: "connection_id"},
		}}
		require.Equal(t, example+"\n", importCommand(def, root))
		got, ok := docImport(def, resourceType, root)
		require.True(t, ok)
		require.Contains(t, got, "```shell\n"+example+"\n```")
		require.Equal(t, "PROJECT/SERVICE_NAME", importID(def, root))
		require.Contains(t, strings.Join(strings.Fields(got), " "),
			`identity = { project = "PROJECT" service_name = "SERVICE_NAME" }`)
	})
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
	require.Contains(t, got, "terraform import aiven_example.example PROJECT/EXAMPLE_ID/REGION")
	require.Contains(t, got, "Terraform 1.12")
	require.Contains(t, got, "identity = {")
	require.Contains(t, got, `project = "PROJECT"`)
	require.Contains(t, got, `example_id = "EXAMPLE_ID"`)
	require.Contains(t, got, `region = "REGION" # Optional`)
	_, ok = docImport(def, datasourceType, root)
	require.False(t, ok)
}
