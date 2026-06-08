package main

import (
	"bytes"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

func TestGenNewResourceRefreshStateDesiredValidation(t *testing.T) {
	// newDef builds the minimum Definition that genNewResource needs to walk into the
	// RefreshState branch: a Resource with RefreshState=true and the desired map under test.
	newDef := func(desired map[string]string) *Definition {
		return &Definition{
			Resource: &SchemaMeta{
				RefreshState:        true,
				RefreshStateDesired: desired,
			},
		}
	}

	t.Run("accepts a desired key that exists and has no enum", func(t *testing.T) {
		item := &Item{
			Properties: map[string]*Item{
				"state": {Name: "state", Type: SchemaTypeString},
			},
		}

		code, err := genNewResource(resourceType, newDef(map[string]string{"state": "ACTIVE"}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `RefreshStateDesired: map[string]string{"state": "ACTIVE"},`)
	})

	t.Run("accepts a desired value that is in the field's enum", func(t *testing.T) {
		item := &Item{
			Properties: map[string]*Item{
				"state": {
					Name: "state",
					Type: SchemaTypeString,
					Enum: []any{"ACTIVE", "APPROVED", "DELETED", "DELETING"},
				},
			},
		}

		code, err := genNewResource(resourceType, newDef(map[string]string{"state": "ACTIVE"}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `RefreshStateDesired: map[string]string{"state": "ACTIVE"},`)
	})

	t.Run("rejects a desired key that does not exist in the schema", func(t *testing.T) {
		item := &Item{
			Properties: map[string]*Item{
				"state": {Name: "state", Type: SchemaTypeString},
			},
		}

		_, err := genNewResource(resourceType, newDef(map[string]string{"missing": "ACTIVE"}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshStateDesired: unknown field "missing"`)
	})

	t.Run("rejects a desired value that is not in the field's enum", func(t *testing.T) {
		item := &Item{
			Properties: map[string]*Item{
				"state": {
					Name: "state",
					Type: SchemaTypeString,
					Enum: []any{"ACTIVE", "APPROVED"},
				},
			},
		}

		_, err := genNewResource(resourceType, newDef(map[string]string{"state": "PENDING"}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshStateDesired: field "state"`)
		require.Contains(t, err.Error(), `desired value "PENDING" is not allowed`)
	})

	t.Run("compares enum values via fmt.Sprint so non-string enums match string desired", func(t *testing.T) {
		// Mirrors adapter.Equal's runtime comparison: fmt.Sprint(7) == "7".
		item := &Item{
			Properties: map[string]*Item{
				"phase": {
					Name: "phase",
					Type: SchemaTypeInteger,
					Enum: []any{1, 7, 42},
				},
			},
		}

		code, err := genNewResource(resourceType, newDef(map[string]string{"phase": "7"}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `RefreshStateDesired: map[string]string{"phase": "7"},`)
	})
}

func renderCode(t *testing.T, code ...jen.Code) string {
	t.Helper()

	file := jen.NewFile("projectvpc")
	for _, c := range code {
		file.Add(c)
	}

	var buf bytes.Buffer
	require.NoError(t, file.Render(&buf))

	return buf.String()
}
