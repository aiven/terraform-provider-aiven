package user_config

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"golang.org/x/exp/maps"
)

// handlePrimitiveTypeProperty is a function that converts a primitive type property to a Terraform schema.
func handlePrimitiveTypeProperty(n string, p map[string]interface{}, t string) map[string]*jen.Statement {
	return map[string]*jen.Statement{n: jen.Values(convertPropertyToSchema(n, p, t))}
}

// handleObjectProperty is a function that converts an object type property to a Terraform schema.
func handleObjectProperty(n string, p map[string]interface{}, t string) map[string]*jen.Statement {
	pa, ok := p["properties"].(map[string]interface{})
	if !ok {
		if it, ok := p["items"].(map[string]interface{}); ok {
			if pa, ok = it["properties"].(map[string]interface{}); !ok {
				return nil
			}
		} else {
			return nil
		}
	}

	r := convertPropertyToSchema(n, p, t)

	s := jen.Map(jen.String()).Op("*").Qual(SchemaPackage, "Schema").Values(convertPropertiesToSchemaMap(pa))

	r[jen.Id("Elem")] = jen.Op("&").Qual(SchemaPackage, "Resource").Values(jen.Dict{
		jen.Id("Schema"): s,
	})

	// TODO: Check if we can access the schema via diff suppression function.
	r[jen.Id("DiffSuppressFunc")] = jen.Qual(SchemaUtilPackage, "EmptyObjectDiffSuppressFuncSkipArrays").Call(s)

	r[jen.Id("MaxItems")] = jen.Lit(1)

	return map[string]*jen.Statement{n: jen.Values(r)}
}

// handleArrayOfPrimitiveTypeProperty is a function that converts an array of primitive type property to a Terraform
// schema.
func handleArrayOfPrimitiveTypeProperty(n string, t string) *jen.Statement {
	r := jen.Dict{
		jen.Id("Type"): jen.Qual(SchemaPackage, t),
	}

	if n == "ip_filter" {
		// TODO: Add ip_filter_object to this sanity check when DiffSuppressFunc is implemented for it.
		r[jen.Id("DiffSuppressFunc")] = jen.Qual(SchemaUtilPackage, "IpFilterValueDiffSuppressFunc")
	}

	return jen.Op("&").Qual(SchemaPackage, "Schema").Values(r)
}

// handleArrayOfAggregateTypeProperty is a function that converts an array of aggregate type property to a Terraform
// schema.
func handleArrayOfAggregateTypeProperty(ip map[string]interface{}) *jen.Statement {
	return jen.Op("&").Qual(SchemaPackage, "Resource").Values(jen.Dict{
		jen.Id("Schema"): jen.Map(jen.String()).Op("*").Qual(SchemaPackage, "Schema").Values(
			convertPropertiesToSchemaMap(ip),
		),
	})
}

// handleArrayProperty is a function that converts an array type property to a Terraform schema.
func handleArrayProperty(n string, p map[string]interface{}, t string) map[string]*jen.Statement {
	ia, ok := p["items"].(map[string]interface{})
	if !ok {
		return nil
	}

	// TODO: Handle this case separately.
	if _, ok := ia["one_of"]; ok {
		return nil
	}

	var e *jen.Statement

	tn, atn := terraformTypes(slicedString(ia["type"]))

	r := make(map[string]*jen.Statement)

	for k, v := range tn {
		an := n
		if len(tn) > 1 {
			an = fmt.Sprintf("%s_%s", n, atn[k])

			// TODO: Remove with the next major version.
			if an == "ip_filter_string" {
				an = "ip_filter"
			}
		}

		if isTerraformTypePrimitive(v) {
			e = handleArrayOfPrimitiveTypeProperty(an, v)
		} else {
			if ipa, ok := ia["properties"].(map[string]interface{}); ok {
				e = handleArrayOfAggregateTypeProperty(ipa)
			} else {
				panic(fmt.Sprintf("could not find properties in an array of aggregate type: %#v", p))
			}
		}

		s := convertPropertyToSchema(n, p, t)

		s[jen.Id("Elem")] = e

		if an == "ip_filter" {
			// TODO: Add ip_filter_object to this sanity check when DiffSuppressFunc is implemented for it.
			s[jen.Id("DiffSuppressFunc")] = jen.Qual(SchemaUtilPackage, "IpFilterArrayDiffSuppressFunc")
		}

		if mi, ok := p["max_items"].(int); ok {
			s[jen.Id("MaxItems")] = jen.Lit(mi)
		}

		r[an] = jen.Values(s)
	}

	return r
}

// handleAggregateTypeProperty is a function that converts an aggregate type property to a Terraform schema.
func handleAggregateTypeProperty(n string, p map[string]interface{}, t string, at string) map[string]*jen.Statement {
	r := make(map[string]*jen.Statement)

	switch at {
	case "object":
		maps.Copy(r, handleObjectProperty(n, p, t))
	case "array":
		maps.Copy(r, handleArrayProperty(n, p, t))
	default:
		panic(fmt.Sprintf("unknown aggregate type: %s", at))
	}

	return r
}
