package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenNewResourceRefreshStateValidation(t *testing.T) {
	attribute := func(value string) *string { return &value }
	newDef := func(refreshState *RefreshStateCondition) *Definition {
		return &Definition{
			Resource: &SchemaMeta{
				RefreshState: refreshState,
			},
		}
	}
	newItem := func(properties map[string]*Item) *Item {
		item := &Item{Properties: properties}
		setParents(item, nil)
		return item
	}

	stringItem := func() *Item {
		return newItem(map[string]*Item{
			"state": {Name: "state", Type: SchemaTypeString, Computed: true},
		})
	}

	t.Run("an absent config disables refresh", func(t *testing.T) {
		code, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{}}, stringItem(), false)
		require.NoError(t, err)
		require.NotContains(t, renderCode(t, code), "RefreshState:")
	})

	t.Run("an empty refreshState enables refresh without conditions", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(&RefreshStateCondition{}), newItem(nil), false)
		require.NoError(t, err)
		require.Contains(t, renderCode(t, code), "&adapter.RefreshStateCondition{}")
	})

	t.Run("generates desired and failed values", func(t *testing.T) {
		item := newItem(map[string]*Item{
			"state": {
				Name:     "state",
				Type:     SchemaTypeString,
				Computed: true,
				Enum:     []any{"ACTIVE", "PENDING_PEER", "ERROR"},
			},
		})

		code, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE", "PENDING_PEER"},
			Failed:  []string{"ERROR"},
		}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "&adapter.RefreshStateCondition{")
		require.Contains(t, got, `Attribute: "state"`)
		require.Contains(t, got, "Desired:")
		require.Contains(t, got, `[]string{"ACTIVE", "PENDING_PEER"}`)
		require.Contains(t, got, "Failed:")
		require.Contains(t, got, `[]string{"ERROR"}`)
	})

	t.Run("generates a custom attribute", func(t *testing.T) {
		item := newItem(map[string]*Item{
			"status": {Name: "status", Type: SchemaTypeString, Computed: true},
		})

		code, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Attribute: attribute("status"),
			Desired:   []string{"ACTIVE"},
		}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `Attribute: "status"`)
		require.Contains(t, got, "Desired:")
		require.Contains(t, got, `[]string{"ACTIVE"}`)
	})

	t.Run("rejects conditions when the default attribute is missing", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE"},
		}), newItem(nil), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshState: unknown attribute "state"`)
	})

	t.Run("rejects configurable fields", func(t *testing.T) {
		for name, field := range map[string]*Item{
			"required": {
				Name:     "state",
				Type:     SchemaTypeString,
				Required: true,
			},
			"optional and computed": {
				Name:     "state",
				Type:     SchemaTypeString,
				Optional: true,
				Computed: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				item := newItem(map[string]*Item{"state": field})
				_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
					Desired: []string{"ACTIVE"},
				}), item, false)
				require.EqualError(t, err, `refreshState: attribute "state" must be computed-only`)
			})
		}
	})

	t.Run("rejects an empty or unused attribute", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Attribute: attribute(""),
			Desired:   []string{"ACTIVE"},
		}), stringItem(), false)
		require.EqualError(t, err, "refreshState: attribute cannot be empty")

		_, err = genNewResource(resourceType, newDef(&RefreshStateCondition{
			Attribute: attribute("status"),
		}), stringItem(), false)
		require.EqualError(t, err, "refreshState: attribute requires desired")
	})

	t.Run("rejects a condition without desired values", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Failed: []string{"ERROR"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `attribute "state" must define at least one desired value`)
	})

	t.Run("rejects an explicitly empty failed list", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE"}, Failed: []string{},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `attribute "state" has an empty failed list`)
	})

	t.Run("rejects desired and failed values outside the field enum", func(t *testing.T) {
		item := newItem(map[string]*Item{
			"state": {
				Name:     "state",
				Type:     SchemaTypeString,
				Computed: true,
				Enum:     []any{"ACTIVE", "APPROVED"},
			},
		})

		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"PENDING"},
		}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshState: attribute "state"`)
		require.Contains(t, err.Error(), `desired value "PENDING" is not allowed`)

		_, err = genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE"}, Failed: []string{"ERROR"},
		}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `failed value "ERROR" is not allowed`)
	})

	t.Run("rejects duplicate and overlapping values", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE", "ACTIVE"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `duplicate desired value "ACTIVE"`)

		_, err = genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE"}, Failed: []string{"ERROR", "ERROR"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `duplicate failed value "ERROR"`)

		_, err = genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"ACTIVE"}, Failed: []string{"ACTIVE"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `value "ACTIVE" cannot be both desired and failed`)
	})

	t.Run("compares enum values via fmt.Sprint so non-string enums match string desired", func(t *testing.T) {
		// Mirrors adapter.Equal's runtime comparison: fmt.Sprint(7) == "7".
		item := newItem(map[string]*Item{
			"state": {
				Name:     "state",
				Type:     SchemaTypeInteger,
				Computed: true,
				Enum:     []any{1, 7, 42},
			},
		})

		code, err := genNewResource(resourceType, newDef(&RefreshStateCondition{
			Desired: []string{"7"}, Failed: []string{"42"},
		}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "Desired:")
		require.Contains(t, got, `[]string{"7"}`)
		require.Contains(t, got, "Failed:")
		require.Contains(t, got, `[]string{"42"}`)
	})

	t.Run("generates refreshStateDelay with refreshState", func(t *testing.T) {
		def := newDef(&RefreshStateCondition{Desired: []string{"ACTIVE"}})
		def.Resource.RefreshStateDelay = 15 * time.Second

		code, err := genNewResource(resourceType, def, stringItem(), false)
		require.NoError(t, err)
		require.Contains(t, renderCode(t, code), `RefreshStateDelay: adapter.MustParseDuration("15s")`)
	})

	t.Run("rejects dependent settings without refreshState", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{
			Resource: &SchemaMeta{RefreshStateDelay: 10},
		}, stringItem(), false)
		require.EqualError(t, err, "refreshStateDelay requires refreshState")

		_, err = genNewResource(resourceType, &Definition{
			Resource: &SchemaMeta{IgnoreAlreadyExists: true},
		}, stringItem(), false)
		require.EqualError(t, err, "ignoreAlreadyExists requires refreshState")
	})
}

func TestSchemaMetaRefreshStateYAML(t *testing.T) {
	decode := func(t *testing.T, source string) (*Definition, error) {
		t.Helper()
		var def Definition
		decoder := yaml.NewDecoder(strings.NewReader(source))
		decoder.KnownFields(true)
		return &def, decoder.Decode(&def)
	}

	t.Run("omitted", func(t *testing.T) {
		def, err := decode(t, "resource: {}\n")
		require.NoError(t, err)
		require.Nil(t, def.Resource.RefreshState)
	})

	t.Run("empty", func(t *testing.T) {
		def, err := decode(t, "resource:\n  refreshState: {}\n")
		require.NoError(t, err)
		require.NotNil(t, def.Resource.RefreshState)
		require.Nil(t, def.Resource.RefreshState.Attribute)
		require.Nil(t, def.Resource.RefreshState.Desired)
		require.Nil(t, def.Resource.RefreshState.Failed)
	})

	t.Run("configured", func(t *testing.T) {
		def, err := decode(t, "resource:\n  refreshState:\n    desired: [ACTIVE]\n    failed: [ERROR]\n")
		require.NoError(t, err)
		require.Nil(t, def.Resource.RefreshState.Attribute)
		require.Equal(t, []string{"ACTIVE"}, def.Resource.RefreshState.Desired)
		require.Equal(t, []string{"ERROR"}, def.Resource.RefreshState.Failed)
	})

	t.Run("custom attribute", func(t *testing.T) {
		def, err := decode(t, "resource:\n  refreshState:\n    attribute: status\n    desired: [ACTIVE]\n")
		require.NoError(t, err)
		require.NotNil(t, def.Resource.RefreshState.Attribute)
		require.Equal(t, "status", *def.Resource.RefreshState.Attribute)
	})

	t.Run("rejects the old nested form", func(t *testing.T) {
		_, err := decode(t, "resource:\n  refreshState:\n    state:\n      desired: [ACTIVE]\n")
		require.ErrorContains(t, err, "field state not found")
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
