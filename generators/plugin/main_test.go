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
