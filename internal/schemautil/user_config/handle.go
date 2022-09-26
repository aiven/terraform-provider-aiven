package user_config

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

// handlePrimitiveTypeProperty is a function that converts a primitive type property to a Terraform schema.
func handlePrimitiveTypeProperty(n string, p map[string]interface{}, t string) *jen.Statement {
	return jen.Values(convertPropertyToSchema(n, p, t))
}

// handleObjectProperty is a function that converts an object type property to a Terraform schema.
func handleObjectProperty(n string, p map[string]interface{}, t string) *jen.Statement {
	pa, ok := p["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	r := convertPropertyToSchema(n, p, t)

	s := jen.Map(jen.String()).Op("*").Qual(SchemaPackage, "Schema").Values(convertPropertiesToSchemaMap(pa))

	r[jen.Id("Elem")] = jen.Op("&").Qual(SchemaPackage, "Resource").Values(jen.Dict{
		jen.Id("Schema"): s,
	})

	// TODO: Check if we can access the schema via diff suppression function.
	r[jen.Id("DiffSuppressFunc")] = jen.Qual(SchemaUtilPackage, "EmptyObjectDiffSuppressFuncSkipArrays").Call(s)

	r[jen.Id("MaxItems")] = jen.Lit(1)

	return jen.Values(r)
}

// handleArrayProperty is a function that converts an array type property to a Terraform schema.
func handleArrayProperty(n string, p map[string]interface{}, t string) *jen.Statement {
	return jen.Values(convertPropertyToSchema(n, p, t))
}

// handleAggregateTypeProperty is a function that converts an aggregate type property to a Terraform schema.
func handleAggregateTypeProperty(n string, p map[string]interface{}, t string, at string) *jen.Statement {
	switch at {
	case "object":
		return handleObjectProperty(n, p, t)
	case "array":
		return handleArrayProperty(n, p, t)
	default:
		panic(fmt.Sprintf("unknown aggregate type: %s", at))
	}
}
