package schemautil

import (
	"reflect"
)

// fillInServiceUserConfigData prepares API compatible service user_config
func fillInServiceUserConfigData(uc interface{}, b map[string]interface{}) map[string]interface{} {
	if uc, ok := uc.([]interface{}); ok {
		if len(uc) == 0 {
			return nil
		}

		if uc[0] == nil {
			return nil
		}

		// TODO: add logic for m3db namespaces similar to ip_filter
		flattenUserConfigIPFilter(uc[0].(map[string]interface{}))

		return walkBlueprint(uc[0].(map[string]interface{}), b)
	}
	return nil
}

// walkBlueprint is a function that recursively traverses two maps c and b
// and generates API compatible user config structure.
// Where:
//   - c is a user config that we get from Terraform state using d.Get(...)
//   - b is a blueprint of the user config API request
//
// It iterates over the data from the Terraform and converts, if necessary,
// to the API-compatible format.
func walkBlueprint(c map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for ck, cv := range c {
		if isUserConfigFieldEmpty(cv) {
			continue
		}

		for bk, bv := range b {
			// if key from the user_config is not equal to blueprint skip it
			if ck != bk {
				continue
			}

			if bv == "primitive" { // primitive type
				m[bk] = cv
			}

			if reflect.TypeOf(bv).Kind() == reflect.Map { // obj type
				bv := bv.(map[string]interface{})
				s := cv.([]interface{})[0]
				if s == nil {
					continue
				}

				m[bk] = walkBlueprint(s.(map[string]interface{}), bv)
			}

			if reflect.TypeOf(bv).Kind() == reflect.Slice { // array of objects
				bv := bv.([]interface{})
				buf := []interface{}{}
				for _, s := range cv.([]interface{}) {
					buf = append(buf, walkBlueprint(s.(map[string]interface{}), bv[0].(map[string]interface{})))
				}

				m[bk] = buf
			}
		}
	}

	return m
}

// isUserConfigFieldEmpty checks if user config filed is empty
func isUserConfigFieldEmpty(cv interface{}) bool {
	if reflect.ValueOf(cv).IsZero() {
		return true
	}

	if _, ok := cv.([]interface{}); ok && len(cv.([]interface{})) == 0 {
		return true
	}

	return false
}

// flattenUserConfigIPFilter convert oneOf ip_filter to API compatible format
func flattenUserConfigIPFilter(c map[string]interface{}) {
	if v, ok := c["ip_filter_string"]; ok && !isUserConfigFieldEmpty(v) {
		c["ip_filter"] = v
	}

	if v, ok := c["ip_filter_object"]; ok && !isUserConfigFieldEmpty(v) {
		c["ip_filter"] = v
	}
}
