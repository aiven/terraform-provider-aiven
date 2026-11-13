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

func TestDocWriteBlockChildrenVariants(t *testing.T) {
	t.Parallel()

	build := func(entity entityType) string {
		settings := &Item{
			Name:     "settings",
			Type:     SchemaTypeObject,
			Optional: entity.isResource(),
			Computed: !entity.isResource(),
			Variants: []*Variant{
				{Title: "Variant 1", Attributes: []string{"v1_only", "nested"}, Required: []string{"v1_only"}},
				{Title: "Variant 2", Attributes: []string{"v2_only", "nested"}, Required: []string{"v2_only"}},
			},
			Properties: map[string]*Item{
				"extra": {
					Name: "extra", Type: SchemaTypeString, Optional: entity.isResource(), Computed: !entity.isResource(),
					Description: "Extra.",
				},
				// Optional in the schema, required by their variants
				"v1_only": {
					Name: "v1_only", Type: SchemaTypeString, Optional: entity.isResource(), Computed: !entity.isResource(),
					Description: "One.",
				},
				"v2_only": {
					Name: "v2_only", Type: SchemaTypeString, Optional: entity.isResource(), Computed: !entity.isResource(),
					Description: "Two.",
				},
				"nested": {
					Name: "nested", Type: SchemaTypeObject, Optional: entity.isResource(), Computed: !entity.isResource(),
					Properties: map[string]*Item{
						"token": {Name: "token", Type: SchemaTypeString, Required: entity.isResource(), Computed: !entity.isResource()},
					},
				},
			},
		}
		settings.Properties["nested"].Parent = settings
		settings.Properties["nested"].Properties["token"].Parent = settings.Properties["nested"]
		for _, p := range settings.Properties {
			if p.Parent == nil {
				p.Parent = settings
			}
		}

		var b strings.Builder
		def := &Definition{Resource: &SchemaMeta{}, Datasource: &SchemaMeta{}}
		require.NoError(t, docWriteBlockChildren(&b, def, entity, settings, []string{"repo"}, false))
		return b.String()
	}

	resource := build(resourceType)
	require.Contains(t, resource, "- `extra` (String) Extra.")
	require.Contains(t, resource, "Set the attributes of exactly one of the variants below.")
	require.Contains(t, resource, "**Variant 1**\n\nRequired:\n\n- `v1_only` (String) One.")
	require.Less(t, strings.Index(resource, "`extra`"), strings.Index(resource, "**Variant 1**"), "the attributes in no variant go first")
	require.Contains(t, resource, "**Variant 2**")
	require.Equal(t, 1, strings.Count(resource, "### Nested Schema for `repo.nested`"))

	anySettings := &Item{
		Name: "settings", Type: SchemaTypeObject, Optional: true, AnyVariant: true,
		Variants: []*Variant{
			{Title: "Variant 1", Attributes: []string{"v1_only"}},
			{Title: "Variant 2", Attributes: []string{"v2_only"}},
		},
		Properties: map[string]*Item{
			"v1_only": {Name: "v1_only", Type: SchemaTypeString, Optional: true},
			"v2_only": {Name: "v2_only", Type: SchemaTypeString, Optional: true},
		},
	}
	for _, p := range anySettings.Properties {
		p.Parent = anySettings
	}
	var b strings.Builder
	require.NoError(t, docWriteBlockChildren(&b, &Definition{Resource: &SchemaMeta{}}, resourceType, anySettings, nil, false))
	require.Contains(t, b.String(), "at least one of the variants below")

	datasource := build(datasourceType)
	require.Contains(t, datasource, "Has the attributes of exactly one of the variants below.")
	require.Contains(t, datasource, "Read-Only:")
	require.Contains(t, datasource, "- `v1_only` (String) One.")
	require.NotContains(t, datasource, "**Variant 1**\n\nRequired:")
}
