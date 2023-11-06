package typeupgrader

import (
	"fmt"
	"strconv"
)

// Map upgrades map values to the specified types.
func Map(valueMap map[string]any, typeRules map[string]string) (err error) {
	for key, targetType := range typeRules {
		valueAsString, ok := valueMap[key].(string)
		if !ok {
			continue
		}

		valueMap[key], err = convert(valueAsString, targetType)
		if err != nil {
			return err
		}
	}

	return nil
}

// Slice upgrades slice values to the specified type.
func Slice(valueSlice []any, targetType string) (err error) {
	for index, value := range valueSlice {
		valueAsString, ok := value.(string)
		if !ok {
			continue
		}

		valueSlice[index], err = convert(valueAsString, targetType)
		if err != nil {
			return err
		}
	}

	return nil
}

// convert converts a value to the specified type.
func convert(value string, targetType string) (convertedValue any, err error) {
	switch targetType {
	case "bool":
		if value == "" {
			value = "false"
		}

		return strconv.ParseBool(value)
	case "int":
		if value == "" {
			value = "0"
		}

		return strconv.Atoi(value)
	case "float":
		if value == "" {
			value = "0"
		}

		return strconv.ParseFloat(value, 64)
	default:
		return nil, fmt.Errorf("unsupported type %q", targetType)
	}
}
