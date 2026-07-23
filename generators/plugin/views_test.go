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

func TestGenNewResourceDeleteStateValidation(t *testing.T) {
	stringItem := func() *Item {
		return &Item{Properties: map[string]*Item{"state": {Name: "state", Type: SchemaTypeString}}}
	}
	enumItem := func(vals ...any) *Item {
		return &Item{Properties: map[string]*Item{
			"state": {Name: "state", Type: SchemaTypeString, Enum: vals},
		}}
	}
	newDef := func(dsd map[string]string) *Definition {
		return &Definition{Resource: &SchemaMeta{DeleteStateDesired: dsd}}
	}

	t.Run("desired conditions generate the poller", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(map[string]string{"state": "DELETED"}), stringItem(), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "DeleteStateOptions{")
		require.Contains(t, got, `map[string]string{"state": "DELETED"}`)
	})

	t.Run("an empty map enables the poller (404-only)", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(map[string]string{}), stringItem(), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "&adapter.DeleteStateOptions{}")
		require.NotContains(t, got, "Desired:")
	})

	t.Run("an absent config disables the poller", func(t *testing.T) {
		code, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{}}, stringItem(), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.NotContains(t, got, "DeleteStateOptions")
	})

	t.Run("accepts a desired value that is in the field's enum", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(map[string]string{"state": "DELETED"}), enumItem("ACTIVE", "DELETING", "DELETED"), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `map[string]string{"state": "DELETED"}`)
	})

	t.Run("rejects a desired value outside the field's enum", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(map[string]string{"state": "deleted"}), enumItem("active", "creating", "deleting"), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `deleteStateDesired: field "state"`)
		require.Contains(t, err.Error(), `desired value "deleted" is not allowed`)
	})

	t.Run("rejects a desired key that does not exist in the schema", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(map[string]string{"missing": "DELETED"}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `deleteStateDesired: unknown field "missing"`)
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
