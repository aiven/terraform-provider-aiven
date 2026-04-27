package main

import (
	"bytes"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

func TestGenGenericViewOperationOnSuccess(t *testing.T) {
	t.Run("delete without response", func(t *testing.T) {
		operation := &Operation{
			ID:        "VpcDelete",
			Type:      OperationDelete,
			OnSuccess: "waitForDeletion",
		}
		definition := &Definition{
			ClientHandler: "vpc",
			Operations:    Operations{operation},
		}

		inPath := definition.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
		item := &Item{
			Properties: map[string]*Item{
				"project": {
					Name:                "project",
					Type:                SchemaTypeString,
					AppearsIn:           inPath,
					IDAttributePosition: 0,
				},
				"project_vpc_id": {
					Name:                "project_vpc_id",
					Type:                SchemaTypeString,
					AppearsIn:           inPath,
					IDAttributePosition: 1,
				},
			},
		}

		code, err := genGenericView(item, definition, OperationDelete)
		require.NoError(t, err)

		got := renderCode(t, code)
		require.Contains(t, got, `err := client.VpcDelete(ctx, d.Get("project").(string), d.Get("project_vpc_id").(string))`)
		require.Contains(t, got, "if err != nil {\n\t\treturn err\n\t}")
		require.Contains(t, got, "return waitForDeletion(ctx, client, d)")
		require.NotContains(t, got, "return client.VpcDelete")
	})

	t.Run("read with response", func(t *testing.T) {
		operation := &Operation{
			ID:        "FooGet",
			Type:      OperationRead,
			OnSuccess: "afterRead",
		}
		definition := &Definition{
			ClientHandler: "foo",
			Operations:    Operations{operation},
		}

		inPath := definition.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
		inResponse := definition.Operations.AppearsInID(operation.ID, operation.Type, ResponseBody)
		item := &Item{
			Properties: map[string]*Item{
				"foo_id": {
					Name:                "foo_id",
					Type:                SchemaTypeString,
					AppearsIn:           inPath,
					IDAttributePosition: 0,
				},
				"name": {
					Name:      "name",
					Type:      SchemaTypeString,
					AppearsIn: inResponse,
				},
			},
		}

		code, err := genGenericView(item, definition, OperationRead)
		require.NoError(t, err)

		got := renderCode(t, code)
		require.Contains(t, got, `rsp, err := client.FooGet(ctx, d.Get("foo_id").(string))`)
		require.Contains(t, got, "err = d.Flatten(rsp)")
		require.Contains(t, got, "if err != nil {\n\t\treturn err\n\t}")
		require.Contains(t, got, "return afterRead(ctx, client, d)")
		require.NotContains(t, got, "return d.Flatten(rsp)")
	})
}

func renderCode(t *testing.T, code jen.Code) string {
	t.Helper()

	file := jen.NewFile("projectvpc")
	file.Add(code)

	var buf bytes.Buffer
	require.NoError(t, file.Render(&buf))

	return buf.String()
}
