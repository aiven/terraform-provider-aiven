package typeupgrader

import (
	"fmt"
	"strconv"
)

// Map upgrades map values to the specified types.
func Map(m map[string]interface{}, rules map[string]string) (err error) {
	for k, t := range rules {
		va, ok := m[k].(string)
		if !ok {
			continue
		}

		m[k], err = convert(va, t)
		if err != nil {
			return err
		}
	}

	return nil
}

// Slice upgrades slice values to the specified type.
func Slice(s []interface{}, t string) (err error) {
	for i, v := range s {
		va, ok := v.(string)
		if !ok {
			continue
		}

		s[i], err = convert(va, t)
		if err != nil {
			return err
		}
	}

	return nil
}

// convert converts a value to the specified type.
func convert(v string, t string) (res interface{}, err error) {
	switch t {
	case "bool":
		if v == "" {
			v = "false"
		}

		return strconv.ParseBool(v)
	case "int":
		if v == "" {
			v = "0"
		}

		return strconv.Atoi(v)
	default:
		return nil, fmt.Errorf("unsupported type %q", t)
	}
}
