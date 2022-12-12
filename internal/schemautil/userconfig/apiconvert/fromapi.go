package apiconvert

import (
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// sliceHasNestedProperties is a function that checks if the given slice has nested properties.
func sliceHasNestedProperties(vr interface{}, va map[string]interface{}) (map[string]interface{}, bool) {
	var res map[string]interface{}

	// rok is the resulting ok value.
	var rok bool

	vra, ok := vr.([]interface{})
	if !ok {
		return res, rok
	}

	for _, v := range vra {
		if p, ok := v.(map[string]interface{}); ok && p != nil {
			rok = true

			break
		}
	}

	if i, ok := va["items"].(map[string]interface{}); ok && rok {
		if p, ok := i["properties"].(map[string]interface{}); ok {
			res = p
		}
	}

	if rok && len(res) == 0 {
		rok = false
	}

	return res, rok
}

// unsettedAPIValue is a function that returns an unsetted value with the given type.
func unsettedAPIValue(t string) interface{} {
	var res interface{}

	switch t {
	// TODO: Uncomment when we use the actual types in the schema.
	//case "boolean":
	//	res = false
	//case "integer":
	//	res = 0
	//case "number":
	//	res = float64(0)
	//case "string":
	//	res = ""
	default:
		res = ""
	case "array":
		res = []interface{}{}
	case "object":
		res = map[string]interface{}{}
	}

	return res
}

// propsFromAPI is a function that converts filled API response properties to Terraform user configuration schema.
func propsFromAPI(n string, r map[string]interface{}, p map[string]interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{}, len(p))

	for k, v := range p {
		va, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s...%s: property is not a map", n, k)
		}

		_, ats, err := userconfig.TerraformTypes(userconfig.SlicedString(va["type"]))
		if err != nil {
			return nil, err
		}

		if len(ats) > 1 {
			return nil, fmt.Errorf("%s...%s: multiple types", n, k)
		}

		t := ats[0]

		vr, ok := r[k]
		if !ok || vr == nil {
			if t == "object" {
				continue
			}

			vr = unsettedAPIValue(t)
		}

		var vrs interface{}

		switch t {
		default:
			switch vra := vr.(type) {
			default:
				// TODO: Drop this when we will be using actual types.
				vrs = fmt.Sprintf("%v", vr)
			case []interface{}:
				var l []interface{}

				if vanp, ok := sliceHasNestedProperties(vr, va); ok {
					for kn, vn := range vra {
						vna, ok := vn.(map[string]interface{})
						if !ok {
							return nil, fmt.Errorf("%s...%s.%d: slice item is not a map", n, k, kn)
						}

						p, err := propsFromAPI(n, vna, vanp)
						if err != nil {
							return nil, err
						}

						l = append(l, p)
					}
				} else {
					l = append(l, vra...)
				}

				// We need to get nested types for the array items, so we can add suffix if needed.
				var nts []string

				if i, ok := va["items"].(map[string]interface{}); ok {
					if oo, ok := i["one_of"].([]interface{}); ok {
						for _, v := range oo {
							if va, ok := v.(map[string]interface{}); ok {
								if vat, ok := va["type"].(string); ok {
									nts = append(nts, vat)
								}
							}
						}
					} else {
						_, nts, err = userconfig.TerraformTypes(userconfig.SlicedString(i["type"]))
						if err != nil {
							return nil, err
						}
					}
				}

				if len(nts) > 1 {
					if len(l) > 0 {
						lf := l[0]

						switch lf.(type) {
						case bool:
							k = fmt.Sprintf("%s_boolean", k)
						case int:
							k = fmt.Sprintf("%s_integer", k)
						case float64:
							k = fmt.Sprintf("%s_number", k)
						case string:
							k = fmt.Sprintf("%s_string", k)
						case []interface{}:
							k = fmt.Sprintf("%s_array", k)
						case map[string]interface{}:
							k = fmt.Sprintf("%s_object", k)
						default:
							return nil, fmt.Errorf("%s...%s: no suffix for given type", n, k)
						}

						// TODO: Remove with the next major version.
						if k == "ip_filter_string" {
							k = "ip_filter"
						}

						// TODO: Remove with the next major version.
						if k == "namespaces_string" {
							k = "namespaces"
						}
					} else {
						for _, v := range nts {
							// TODO: Inline with the next major version.
							tk := fmt.Sprintf("%s_%s", k, v)

							// TODO: Remove with the next major version.
							if tk == "ip_filter_string" {
								tk = "ip_filter"
							}

							// TODO: Remove with the next major version.
							if tk == "namespaces_string" {
								tk = "namespaces"
							}

							res[tk] = l
						}

						continue
					}
				}

				vrs = l
			}
		case "object":
			vra, ok := vr.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s...%s: representation value is not a map", n, k)
			}

			nv, ok := va["properties"]
			if !ok {
				return nil, fmt.Errorf("%s...%s: properties key not found", n, k)
			}

			nva, ok := nv.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s...%s: properties value is not a map", n, k)
			}

			p, err := propsFromAPI(n, vra, nva)
			if err != nil {
				return nil, err
			}

			vrs = []map[string]interface{}{p}
		}

		// TODO: Remove when this is fixed in front end.
		if vrs != nil && k == "ip_filter_object" {
			vrsa, ok := vrs.([]interface{})
			if !ok {
				return nil, fmt.Errorf("%s...%s: ip_filter_object value is not []interface{}", n, k)
			}

			var cif []interface{}

			nde := false

			for _, v := range vrsa {
				va, ok := v.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf(
						"%s...%s: ip_filter_object value is not []map[string]interface{}", n, k,
					)
				}

				vda, ok := va["description"].(string)
				if !ok {
					return nil, fmt.Errorf("%s...%s: description value is not a string", n, k)
				}

				if vda != "" {
					nde = true
				}

				vna, ok := va["network"].(string)
				if !ok {
					return nil, fmt.Errorf("%s...%s: network value is not a string", n, k)
				}

				cif = append(cif, vna)
			}

			if !nde {
				k = "ip_filter"

				vrs = cif
			}
		}

		res[userconfig.EncodeKey(k)] = vrs
	}

	return res, nil
}

// FromAPI is a function that converts filled API response to Terraform user configuration schema.
func FromAPI(st userconfig.SchemaType, n string, r map[string]interface{}) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	if len(r) == 0 {
		return res, nil
	}

	p, err := props(st, n)
	if err != nil {
		return nil, err
	}

	pa, err := propsFromAPI(n, r, p)
	if err != nil {
		return nil, err
	}

	res = append(res, pa)

	return res, nil
}
