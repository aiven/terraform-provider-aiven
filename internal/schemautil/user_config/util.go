package user_config

import "fmt"

// terraformTypes is a function that converts schema representation types to Terraform types.
func terraformTypes(t []string) ([]string, []string) {
	r, ar := []string{}, []string{}

	for _, v := range t {
		switch v {
		case "null":
			// TODO: We should probably handle this case.
			//  This is a special case where the value can be null.
			//  There should be a default value set for this case.
			continue
		case "boolean":
			r = append(r, "TypeBool")
		case "integer":
			r = append(r, "TypeInt")
		case "number":
			r = append(r, "TypeFloat")
		case "string":
			r = append(r, "TypeString")
		case "array", "object":
			r = append(r, "TypeSet")
		default:
			panic(fmt.Sprintf("unknown type: %s", v))
		}

		ar = append(ar, v)
	}

	return r, ar
}

// mustStringSlice is a function that converts an interface to a slice of strings.
func mustStringSlice(v interface{}) []string {
	va, ok := v.([]interface{})
	if !ok {
		panic(fmt.Sprintf("not a slice: %#v", v))
	}

	r := make([]string, len(va))

	for k, v := range va {
		va, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("value is not a string: %#v", v))
		}

		r[k] = va
	}

	return r
}

// slicedString is a function that accepts a string or a slice of strings and returns a slice of strings.
func slicedString(v interface{}) []string {
	va, ok := v.([]interface{})
	if ok {
		return mustStringSlice(va)
	}

	vsa, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("value is not a string or a slice of strings: %#v", v))
	}

	return []string{vsa}
}
