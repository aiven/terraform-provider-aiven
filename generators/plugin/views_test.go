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
	enumItem := func(field string, vals ...any) *Item {
		return &Item{Properties: map[string]*Item{
			field: {Name: field, Type: SchemaTypeString, Enum: vals},
		}}
	}
	newDef := func(ds *DeleteStateMeta) *Definition {
		return &Definition{Resource: &SchemaMeta{DeleteState: ds}}
	}

	t.Run("generates the full poller for a configured field", func(t *testing.T) {
		code, err := genNewResource(resourceType, newDef(&DeleteStateMeta{
			Field:         "status",
			Desired:       "DELETED",
			FailureStates: []string{"FAILED", "CANCELLED"},
		}), enumItem("status", "ACTIVE", "DELETED", "FAILED", "CANCELLED"), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "DeleteStateOptions{")
		require.Regexp(t, `Desired:\s+"DELETED"`, got)
		require.Contains(t, got, `FailureStates: []string{"FAILED", "CANCELLED"}`)
		require.Regexp(t, `Observe:\s+deleteStateObserve`, got)
	})

	t.Run("configuration without a field waits for 404", func(t *testing.T) {
		code, err := genNewResource(
			resourceType,
			newDef(&DeleteStateMeta{}),
			&Item{Properties: map[string]*Item{}},
			false,
		)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "DeleteStateOptions{")
		require.Contains(t, got, `Observe: deleteStateObserve`)
		require.NotContains(t, got, "Desired:")
	})

	t.Run("an absent config disables the poller", func(t *testing.T) {
		code, err := genNewResource(resourceType, &Definition{Resource: &SchemaMeta{}}, enumItem("state"), false)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.NotContains(t, got, "DeleteStateOptions")
	})

	t.Run("rejects invalid state configuration", func(t *testing.T) {
		tests := []struct {
			name string
			meta *DeleteStateMeta
			item *Item
			want string
		}{
			{
				name: "missing field option",
				meta: &DeleteStateMeta{Desired: "DELETED"},
				item: enumItem("state", "DELETED"),
				want: "field is required",
			},
			{
				name: "field absent from schema",
				meta: &DeleteStateMeta{Field: "status", Desired: "DELETED"},
				item: enumItem("state", "DELETED"),
				want: `field "status" is not present in schema`,
			},
			{
				name: "desired outside enum",
				meta: &DeleteStateMeta{Field: "state", Desired: "deleted"},
				item: enumItem("state", "active", "deleting"),
				want: `desired value "deleted" is not allowed`,
			},
			{
				name: "failure outside enum",
				meta: &DeleteStateMeta{Field: "state", FailureStates: []string{"FAILED"}},
				item: enumItem("state", "ACTIVE", "DELETED"),
				want: `failure state "FAILED" is not allowed`,
			},
			{
				name: "desired is also a failure",
				meta: &DeleteStateMeta{Field: "state", Desired: "DELETED", FailureStates: []string{"DELETED"}},
				item: enumItem("state", "DELETED"),
				want: `failure state "DELETED" must not also be desired`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := genNewResource(resourceType, newDef(tt.meta), tt.item, false)
				require.ErrorContains(t, err, tt.want)
			})
		}
	})
}

func TestGenDeleteStateObserve(t *testing.T) {
	newDefinition := func(deleteState *DeleteStateMeta, operations ...*Operation) (*Definition, *Item) {
		def := &Definition{
			ClientHandler: "widget",
			Resource:      &SchemaMeta{DeleteState: deleteState},
			Operations:    operations,
		}
		item := &Item{Properties: map[string]*Item{}}
		if len(operations) > 0 {
			read := operations[0]
			item.Properties["project"] = &Item{
				Parent:              item,
				Name:                "project",
				JSONName:            "project",
				Type:                SchemaTypeString,
				IDAttributePosition: 0,
				AppearsIn:           def.Operations.AppearsInID(read.ID, read.Type, PathParameter),
			}
			if deleteState.Field != "" {
				item.Properties[deleteState.Field] = &Item{
					Parent:    item,
					Name:      deleteState.Field,
					JSONName:  deleteState.Field,
					Type:      SchemaTypeString,
					AppearsIn: def.Operations.AppearsInID(read.ID, read.Type, ResponseBody),
				}
			}
		}
		return def, item
	}

	t.Run("reads a renamed field through the response wrapper", func(t *testing.T) {
		read := &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}, ResultKeyField: "Wrapper"}
		def, item := newDefinition(&DeleteStateMeta{Field: "status", Desired: "DELETED"}, read)
		item.Properties["status"].JSONName = "backend_status"

		code, err := genDeleteStateObserve(def, item)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `client.WidgetGet(ctx, d.Get("project").(string))`)
		require.Contains(t, got, `return rsp.Wrapper.BackendStatus, nil`)
		require.NotContains(t, got, "Flatten")
		require.NotContains(t, got, "readView")
	})

	t.Run("empty desired observes existence without requiring state", func(t *testing.T) {
		read := &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}}
		def, item := newDefinition(&DeleteStateMeta{}, read)

		code, err := genDeleteStateObserve(def, item)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, `_, err := client.WidgetGet(ctx, d.Get("project").(string))`)
		require.Contains(t, got, `return nil, err`)
		require.NotContains(t, got, ".Status")
	})

	t.Run("requires the canonical get to return the configured field", func(t *testing.T) {
		read := &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}}
		def, item := newDefinition(&DeleteStateMeta{Field: "status", FailureStates: []string{"FAILED"}}, read)

		item.Properties["status"].AppearsIn = 0
		_, err := genDeleteStateObserve(def, item)
		require.EqualError(t, err, `canonical read operation "WidgetGet" does not return field "status"`)
	})

	t.Run("ignores datasource lookup reads", func(t *testing.T) {
		read := &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}}
		lookup := &Operation{
			ID:               "WidgetList",
			Type:             OperationRead,
			DatasourceLookup: true,
			Response:         &OASchema{},
		}
		disabled := &Operation{ID: "WidgetReadWithRequest", Type: OperationRead, DisableView: true, Response: &OASchema{}}
		def, item := newDefinition(&DeleteStateMeta{Field: "state", Desired: "DELETED"}, read, lookup, disabled)

		code, err := genDeleteStateObserve(def, item)
		require.NoError(t, err)
		got := renderCode(t, code)
		require.Contains(t, got, "client.WidgetGet")
		require.NotContains(t, got, "client.WidgetList")
		require.NotContains(t, got, "client.WidgetReadWithRequest")
	})

	t.Run("rejects unsupported canonical read shapes", func(t *testing.T) {
		tests := []struct {
			name string
			read *Operation
			want string
		}{
			{
				name: "request body",
				read: &Operation{ID: "WidgetGet", Type: OperationRead, Request: &OASchema{}, Response: &OASchema{}},
				want: "has a request body",
			},
			{
				name: "resultToKey",
				read: &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}, ResultToKey: "widget"},
				want: "wraps its response with resultToKey",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				def, item := newDefinition(&DeleteStateMeta{Field: "state", Desired: "DELETED"}, tt.read)
				_, err := genDeleteStateObserve(def, item)
				require.ErrorContains(t, err, tt.want)
			})
		}
	})

	t.Run("requires one canonical read", func(t *testing.T) {
		def, item := newDefinition(&DeleteStateMeta{Field: "state", Desired: "DELETED"})
		_, err := genDeleteStateObserve(def, item)
		require.EqualError(t, err, "deleteState requires exactly one canonical read operation, got 0")

		first := &Operation{ID: "WidgetGet", Type: OperationRead, Response: &OASchema{}}
		second := &Operation{ID: "WidgetGetAgain", Type: OperationRead, Response: &OASchema{}}
		def, item = newDefinition(&DeleteStateMeta{Field: "state", Desired: "DELETED"}, first, second)
		_, err = genDeleteStateObserve(def, item)
		require.EqualError(t, err, "deleteState requires exactly one canonical read operation, got 2")
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
