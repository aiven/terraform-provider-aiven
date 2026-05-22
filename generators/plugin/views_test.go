package main

import (
	"bytes"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

func TestGenGenericViewOperationWaitForDeletion(t *testing.T) {
	operation := &Operation{
		ID:              "VpcDelete",
		Type:            OperationDelete,
		WaitForDeletion: true,
	}
	definition := &Definition{
		ClientHandler: "vpc",
		Operations:    Operations{operation},
	}

	inPath := definition.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
	inResponse := definition.Operations.AppearsInID(operation.ID, operation.Type, ResponseBody)
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
			"state": {
				Name:      "state",
				Type:      SchemaTypeString,
				AppearsIn: inResponse,
			},
		},
	}

	code, err := genGenericView(definition, item, OperationDelete)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, `_, err := client.VpcDelete(ctx, d.Get("project").(string), d.Get("project_vpc_id").(string))`)
	require.Contains(t, got, "if err != nil {\n\t\treturn err\n\t}")
	require.Contains(t, got, "return waitForDeletion(ctx, client, d)")
}

func TestGenPopulateIDFieldsFromID(t *testing.T) {
	operation := &Operation{
		ID:   "VpcDelete",
		Type: OperationDelete,
	}
	definition := &Definition{
		typeName:            "aiven_project_vpc",
		Resource:            &SchemaMeta{},
		ClientHandler:       "vpc",
		IDAttributeComposed: IDAttribute{Fields: []string{"project", "project_vpc_id"}, PopulateFieldsFromID: true},
		Operations:          Operations{operation},
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

	view, err := genGenericView(definition, item, OperationDelete)
	require.NoError(t, err)

	got := renderCode(t, view, genPopulateIDFieldsFromID(definition))
	require.Contains(t, got, "if err := populateIDFieldsFromID(d); err != nil")
	require.Contains(t, got, "schemautil.SplitResourceID(d.ID(), len(fields))")
	require.Contains(t, got, `fmt.Errorf("invalid aiven_project_vpc id %q: %w", d.ID(), err)`)
	require.Contains(t, got, "d.Set(field, chunks[i])")
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
