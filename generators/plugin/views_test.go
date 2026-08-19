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
	newDef := func(meta SchemaMeta) *Definition {
		return &Definition{Resource: &meta}
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

	t.Run("refreshStateExists enables refresh without conditions", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(SchemaMeta{RefreshStateExists: true}), newItem(nil), false)
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

		code, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE", "PENDING_PEER"},
			RefreshStateFailed:  []string{"ERROR"},
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

	t.Run("generates a custom state field", func(t *testing.T) {
		item := newItem(map[string]*Item{
			"status": {Name: "status", Type: SchemaTypeString, Computed: true},
		})

		code, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "status",
			RefreshStateDesired: []string{"ACTIVE"},
		}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `Attribute: "status"`)
		require.Contains(t, got, "Desired:")
		require.Contains(t, got, `[]string{"ACTIVE"}`)
	})

	t.Run("rejects conditions without stateAttribute", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			RefreshStateDesired: []string{"ACTIVE"},
		}), stringItem(), false)
		require.EqualError(t, err, "stateAttribute is required")
	})

	t.Run("rejects an unknown stateAttribute", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
		}), newItem(nil), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `stateAttribute "state" is not present in schema`)
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
				_, err := genNewResource(resourceType, newDef(SchemaMeta{
					StateAttribute:      "state",
					RefreshStateDesired: []string{"ACTIVE"},
				}), item, false)
				require.EqualError(t, err, `stateAttribute "state" must be computed-only`)
			})
		}
	})

	t.Run("rejects an unused stateAttribute", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:     "state",
			RefreshStateExists: true,
		}), stringItem(), false)
		require.EqualError(t, err, "stateAttribute requires refreshStateDesired or deleteStateDesired")
	})

	t.Run("rejects a condition without desired values", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:     "state",
			RefreshStateFailed: []string{"ERROR"},
		}), stringItem(), false)
		require.EqualError(t, err, "refreshStateFailed requires refreshStateDesired")
	})

	t.Run("rejects an empty desired list", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			RefreshStateDesired: []string{},
		}), stringItem(), false)
		require.EqualError(t, err, "refreshStateDesired must not be empty; use refreshStateExists")
	})

	t.Run("rejects combining exists with desired", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateExists:  true,
			RefreshStateDesired: []string{"ACTIVE"},
		}), stringItem(), false)
		require.EqualError(t, err, "refreshStateExists is implied by refreshStateDesired; omit refreshStateExists")
	})

	t.Run("rejects an explicitly empty failed list", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
			RefreshStateFailed:  []string{},
		}), stringItem(), false)
		require.EqualError(t, err, "refreshStateFailed must not be empty")
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

		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"PENDING"},
		}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshStateDesired: "state"`)
		require.Contains(t, err.Error(), `value "PENDING" is not allowed`)

		_, err = genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
			RefreshStateFailed:  []string{"ERROR"},
		}), item, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `refreshStateFailed: "state"`)
		require.Contains(t, err.Error(), `value "ERROR" is not allowed`)
	})

	t.Run("rejects duplicate and overlapping values", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE", "ACTIVE"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `duplicate value "ACTIVE"`)

		_, err = genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
			RefreshStateFailed:  []string{"ERROR", "ERROR"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `duplicate value "ERROR"`)

		_, err = genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
			RefreshStateFailed:  []string{"ACTIVE"},
		}), stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `value "ACTIVE" cannot be both refreshStateDesired and refreshStateFailed`)
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

		code, err := genNewResource(resourceType, newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"7"},
			RefreshStateFailed:  []string{"42"},
		}), item, false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "Desired:")
		require.Contains(t, got, `[]string{"7"}`)
		require.Contains(t, got, "Failed:")
		require.Contains(t, got, `[]string{"42"}`)
	})

	t.Run("generates refreshStateDelay with refreshStateDesired", func(t *testing.T) {
		def := newDef(SchemaMeta{
			StateAttribute:      "state",
			RefreshStateDesired: []string{"ACTIVE"},
			RefreshStateDelay:   15 * time.Second,
		})

		code, err := genNewResource(resourceType, def, stringItem(), false)
		require.NoError(t, err)
		require.Contains(t, renderCode(t, code), `RefreshStateDelay: adapter.MustParseDuration("15s")`)
	})

	t.Run("rejects dependent settings without a refresh", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{
			Resource: &SchemaMeta{RefreshStateDelay: 10},
		}, stringItem(), false)
		require.EqualError(t, err, "refreshStateDelay requires refreshStateExists or refreshStateDesired")

		_, err = genNewResource(resourceType, &Definition{
			Resource: &SchemaMeta{IgnoreAlreadyExists: true},
		}, stringItem(), false)
		require.EqualError(t, err, "ignoreAlreadyExists requires refreshStateExists or refreshStateDesired")
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
		require.False(t, def.Resource.RefreshStateExists)
		require.Nil(t, def.Resource.RefreshStateDesired)
		require.Nil(t, def.Resource.RefreshStateFailed)
		require.Empty(t, def.Resource.StateAttribute)
	})

	t.Run("exists", func(t *testing.T) {
		def, err := decode(t, "resource:\n  refreshStateExists: true\n")
		require.NoError(t, err)
		require.True(t, def.Resource.RefreshStateExists)
		require.Nil(t, def.Resource.RefreshStateDesired)
	})

	t.Run("configured", func(t *testing.T) {
		def, err := decode(t, "resource:\n  stateAttribute: state\n  refreshStateDesired: [ACTIVE]\n  refreshStateFailed: [ERROR]\n")
		require.NoError(t, err)
		require.Equal(t, "state", def.Resource.StateAttribute)
		require.Equal(t, []string{"ACTIVE"}, def.Resource.RefreshStateDesired)
		require.Equal(t, []string{"ERROR"}, def.Resource.RefreshStateFailed)
	})

	t.Run("custom state field", func(t *testing.T) {
		def, err := decode(t, "resource:\n  stateAttribute: status\n  refreshStateDesired: [ACTIVE]\n")
		require.NoError(t, err)
		require.Equal(t, "status", def.Resource.StateAttribute)
	})

	t.Run("rejects the old nested form", func(t *testing.T) {
		_, err := decode(t, "resource:\n  refreshState:\n    attribute: state\n    desired: [ACTIVE]\n")
		require.ErrorContains(t, err, "field refreshState not found")
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
	newDef := func(desired []string) *Definition {
		meta := SchemaMeta{DeleteStateDesired: desired}
		if len(desired) > 0 {
			meta.StateAttribute = "state"
		}
		return &Definition{Resource: &meta}
	}

	t.Run("desired conditions generate the poller", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef([]string{"DELETED"}), stringItem(), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "DeleteStateOptions{")
		require.Contains(t, got, `Attribute: "state"`)
		require.Contains(t, got, `[]string{"DELETED"}`)
	})

	t.Run("deleteStateGone enables the poller (404-only)", func(t *testing.T) {
		code, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{DeleteStateGone: true}}, stringItem(), false)
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
		code, err := genNewResource(resourceType, newDef([]string{"DELETED"}), enumItem("ACTIVE", "DELETING", "DELETED"), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `[]string{"DELETED"}`)
	})

	t.Run("rejects a desired value outside the field's enum", func(t *testing.T) {
		_, err := genNewResource(resourceType, newDef([]string{"deleted"}), enumItem("active", "creating", "deleting"), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `deleteStateDesired: "state"`)
		require.Contains(t, err.Error(), `value "deleted" is not allowed`)
	})

	t.Run("rejects a stateAttribute that does not exist in the schema", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{
			StateAttribute:     "missing",
			DeleteStateDesired: []string{"DELETED"},
		}}, stringItem(), false)
		require.Error(t, err)
		require.Contains(t, err.Error(), `stateAttribute "missing" is not present in schema`)
	})

	t.Run("rejects delete values without stateAttribute", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{
			DeleteStateDesired: []string{"DELETED"},
		}}, stringItem(), false)
		require.EqualError(t, err, "stateAttribute is required")
	})

	t.Run("rejects an empty desired list", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{
			DeleteStateDesired: []string{},
		}}, stringItem(), false)
		require.EqualError(t, err, "deleteStateDesired must not be empty; use deleteStateGone")
	})

	t.Run("rejects combining gone with desired", func(t *testing.T) {
		_, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{
			StateAttribute:     "state",
			DeleteStateGone:    true,
			DeleteStateDesired: []string{"DELETED"},
		}}, stringItem(), false)
		require.EqualError(t, err, "deleteStateGone is implied by deleteStateDesired; omit deleteStateGone")
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
