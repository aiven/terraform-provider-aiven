package typeupgrader

import (
	"strconv"
)

// Map upgrades map values to the specified types.
func Map(m map[string]interface{}, rules map[string]string) (err error) {
	for k, t := range rules {
		va, ok := m[k].(string)
		if !ok {
			continue
		}

		switch t {
		case "bool":
			if va == "" {
				va = "false"
			}

			m[k], err = strconv.ParseBool(va)
			if err != nil {
				return err
			}
		case "int":
			if va == "" {
				va = "0"
			}

			m[k], err = strconv.Atoi(va)
			if err != nil {
				return err
			}
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

		switch t {
		case "int":
			if va == "" {
				va = "0"
			}

			s[i], err = strconv.Atoi(va)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
