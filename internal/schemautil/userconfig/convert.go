//nolint:unused
package userconfig

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

// convertPropertyToSchema is a function that converts a property to a Terraform schema.
func convertPropertyToSchema(n string, p map[string]interface{}, t string, ad bool, ireq bool) jen.Dict {
	r := jen.Dict{
		jen.Id("Type"): jen.Qual(SchemaPackage, t),
	}

	if ad {
		id, d := descriptionForProperty(p, t)

		r[jen.Id("Description")] = jen.Lit(d)

		if id {
			r[jen.Id("Deprecated")] = jen.Lit("Usage of this field is discouraged.")
		}
	}

	if ireq {
		r[jen.Id("Required")] = jen.Lit(true)
	} else {
		r[jen.Id("Optional")] = jen.Lit(true)
	}

	if d, ok := p["default"]; ok && isTerraformTypePrimitive(t) {
		r[jen.Id("Default")] = jen.Lit(d)
	}

	if co, ok := p["create_only"]; ok && co.(bool) {
		r[jen.Id("ForceNew")] = jen.Lit(true)
	}

	if strings.Contains(n, "api_key") || strings.Contains(n, "password") {
		r[jen.Id("Sensitive")] = jen.Lit(true)
	}

	// TODO: Generate validation rules for generated schema properties, also validate that value is within enum values.

	return r
}

// convertPropertiesToSchemaMap is a function that converts a map of properties to a map of Terraform schemas.
func convertPropertiesToSchemaMap(p map[string]interface{}, req map[string]struct{}) (jen.Dict, error) {
	r := make(jen.Dict, len(p))

	for k, v := range p {
		va, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		ts, ats, err := TerraformTypes(SlicedString(va["type"]))
		if err != nil {
			return nil, err
		}

		if len(ts) > 1 {
			return nil, fmt.Errorf("multiple types for %s", k)
		}

		t, at := ts[0], ats[0]

		_, ireq := req[k]

		var s map[string]*jen.Statement

		if isTerraformTypePrimitive(t) {
			s = handlePrimitiveTypeProperty(k, va, t, ireq)
		} else {
			s, err = handleAggregateTypeProperty(k, va, t, at)
			if err != nil {
				return nil, err
			}
		}

		if s == nil {
			continue
		}

		for kn, vn := range s {
			r[jen.Lit(EncodeKey(kn))] = vn
		}
	}

	return r, nil
}
