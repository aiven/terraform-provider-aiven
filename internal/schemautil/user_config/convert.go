package user_config

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

// convertPropertyToSchema is a function that converts a property to a Terraform schema.
func convertPropertyToSchema(n string, p map[string]interface{}, t string) jen.Dict {
	r := jen.Dict{
		jen.Id("Type"):     jen.Qual(SchemaPackage, t),
		jen.Id("Optional"): jen.Lit(true),
	}

	if d, ok := p["description"]; ok {
		k := "Description"

		if strings.Contains(strings.ToLower(d.(string)), "deprecated") {
			k = "Deprecated"
		}

		r[jen.Id(k)] = jen.Lit(d)
	} else {
		if title, ok := p["title"]; ok {
			r[jen.Id("Description")] = jen.Lit(title)
		}
	}

	if co, ok := p["create_only"]; ok && co.(bool) {
		r[jen.Id("DiffSuppressFunc")] = jen.Qual(SchemaUtilPackage, "CreateOnlyDiffSuppressFunc")
	}

	if strings.Contains(n, "api_key") || strings.Contains(n, "password") {
		r[jen.Id("Sensitive")] = jen.Lit(true)
	}

	// TODO: Generate validation rules for generated schema properties, also validate that value is within enum values.

	return r
}

// convertPropertiesToSchemaMap is a function that converts a map of properties to a map of Terraform schemas.
func convertPropertiesToSchemaMap(p map[string]interface{}) jen.Dict {
	r := make(jen.Dict, len(p))

	for k, v := range p {
		va, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		t, at := terraformTypes(slicedString(va["type"]))

		for kn, vn := range t {
			var s *jen.Statement

			switch vn {
			case "TypeBool", "TypeInt", "TypeFloat", "TypeString":
				s = handlePrimitiveTypeProperty(k, va, vn)
			default:
				s = handleAggregateTypeProperty(k, va, vn, at[kn])
			}

			if s == nil {
				continue
			}

			r[jen.Lit(k)] = s
		}
	}

	return r
}
