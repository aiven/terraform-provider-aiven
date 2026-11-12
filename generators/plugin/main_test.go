package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestListDefinitionFilesEmptyInclude(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".empty.yml"), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aiven_foo.yml"), []byte("include: .empty.yml\nlegacyTimeouts: true\n"), 0o644))

	_, err := listDefinitionFiles(dir)
	assert.EqualError(t, err, `definition "aiven_foo.yml": include ".empty.yml" must contain a YAML map`)
}

func TestMergeYAML(t *testing.T) {
	var base, def map[string]any
	assert.NoError(t, yaml.Unmarshal([]byte(`
remove: [acl]
legacyTimeouts: true
schema:
  plan:
    description: Base plan
    required: true
`), &base))
	assert.NoError(t, yaml.Unmarshal([]byte(`
remove: [topics]
schema:
  plan:
    description: Kafka plan
  kafka:
    computed: true
`), &def))

	expected := map[string]any{
		"remove":         []any{"acl", "topics"},
		"legacyTimeouts": true,
		"schema": map[string]any{
			"plan":  map[string]any{"description": "Kafka plan", "required": true},
			"kafka": map[string]any{"computed": true},
		},
	}
	assert.Equal(t, expected, mergeYAML(base, def))
}

func TestFromSchemaHoist(t *testing.T) {
	t.Parallel()
	schema := &OASchema{
		Type: SchemaTypeObject,
		Properties: map[string]*OASchema{
			"access_control": {
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"pg_allow_replication": {Type: SchemaTypeBoolean},
					"redis_acl_keys":       {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
				},
			},
			"username": {Type: SchemaTypeString},
		},
	}
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: make(map[string]*Item)}
	scope := &Scope{Definition: &Definition{
		Hoist:  map[string][]string{"access_control": {"*"}},
		Remove: []string{"access_control/redis_*"},
	}}
	require.NoError(t, fromSchema(scope, &Operation{}, root, schema, ReadHandler))
	assert.Contains(t, root.Properties, "pg_allow_replication")
	assert.NotContains(t, root.Properties, "redis_acl_keys")
	assert.NotContains(t, root.Properties, "access_control")
	assert.Equal(t, "access_control", root.Properties["pg_allow_replication"].HoistedFrom)
}

func TestAddPropertyHoistCollision(t *testing.T) {
	t.Parallel()
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{
		"foo": {Name: "foo", HoistedFrom: ""},
	}}
	scope := &Scope{Definition: &Definition{}}
	err := addProperty(scope, root, &Item{Name: "foo", HoistedFrom: "access_control"})
	require.ErrorContains(t, err, "collides with root field")
}

func TestPruneCrossGroupFieldsHoistedJSONPath(t *testing.T) {
	t.Parallel()
	def := &Definition{typeName: "aiven_pg_user"}
	root := &Item{Name: "root", Properties: map[string]*Item{}}
	child := &Item{
		Parent:      root,
		Description: "valkey keys",
		JSONName:    "valkey_acl_keys",
		HoistedFrom: "access_control",
	}
	root.Properties["valkey_acl_keys"] = child
	require.True(t, pruneCrossGroupFields(def, root, serviceTypes))
	require.Equal(t, "valkey keys", def.prunedFields["access_control/valkey_acl_keys"])
}
