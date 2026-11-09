package main

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestDefinitionsMatchSchema keeps definitions/.schema.yml, which editors validate against, in sync with the generator.
func TestDefinitionsMatchSchema(t *testing.T) {
	dir := filepath.Join("..", "..", "definitions")
	data, err := os.ReadFile(filepath.Join(dir, ".schema.yml"))
	require.NoError(t, err)

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	require.NoError(t, c.AddResource("schema.json", yamlToJSON(t, data)))
	schema, err := c.Compile("schema.json")
	require.NoError(t, err)

	files, err := filepath.Glob(filepath.Join(dir, "aiven_*.yml"))
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, f := range files {
		name := filepath.Base(f)
		t.Run(name, func(t *testing.T) {
			data, err := readDefinition(dir, name)
			require.NoError(t, err)
			require.NoError(t, schema.Validate(yamlToJSON(t, data)))
		})
	}
}

// TestSchemaMatchesStructs checks that .schema.yml documents the keys the generator decodes, and no others.
func TestSchemaMatchesStructs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "definitions", ".schema.yml"))
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, yaml.Unmarshal(data, &schema))

	props := func(path ...string) []string {
		node := schema
		for _, p := range path {
			node = node[p].(map[string]any)
		}
		return slices.Collect(maps.Keys(node["properties"].(map[string]any)))
	}

	// include is merged before decoding, see readDefinition
	topLevel := slices.DeleteFunc(props(), func(k string) bool { return k == "include" })
	meta := lo.Uniq(append(props("properties", "resource"), props("properties", "datasource")...))
	cases := []struct {
		name  string
		typ   reflect.Type
		props []string
	}{
		{name: "definition", typ: reflect.TypeFor[Definition](), props: topLevel},
		{name: "resource and datasource", typ: reflect.TypeFor[SchemaMeta](), props: meta},
		{name: "operation", typ: reflect.TypeFor[Operation](), props: props("definitions", "Operation")},
		{name: "pagination", typ: reflect.TypeFor[Pagination](), props: props("definitions", "Operation", "properties", "pagination")},
		{name: "error rule", typ: reflect.TypeFor[ErrorRule](), props: props("definitions", "Operation", "properties", "errors", "items")},
		{name: "attribute", typ: reflect.TypeFor[Item](), props: props("definitions", "Schema")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.ElementsMatch(t, decodedKeys(t, c.typ), c.props)
		})
	}
}

// decodedKeys returns the YAML keys of the exported fields, without runtime ones: "-" and "_" prefixed.
func decodedKeys(t *testing.T, typ reflect.Type) []string {
	t.Helper()
	var keys []string
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, ok := f.Tag.Lookup("yaml")
		require.True(t, ok, "%s.%s has no yaml tag", typ.Name(), f.Name)
		key := strings.Split(tag, ",")[0]
		if key != "-" && !strings.HasPrefix(key, "_") {
			keys = append(keys, key)
		}
	}
	return keys
}

// yamlToJSON decodes YAML into the JSON values the validator expects.
func yamlToJSON(t *testing.T, data []byte) any {
	t.Helper()
	var doc any
	require.NoError(t, yaml.Unmarshal(data, &doc))
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	v, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	require.NoError(t, err)
	return v
}
