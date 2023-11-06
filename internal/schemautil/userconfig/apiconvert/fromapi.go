package apiconvert

import (
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// hasNestedProperties is a function that returns a map of nested properties and a boolean indicating whether the given
// value has nested properties.
func hasNestedProperties(
	valueReceived any,
	valueAttributes map[string]any,
) (map[string]any, bool) {
	var properties map[string]any
	var resultOk bool

	valueReceivedAsArray, isArray := valueReceived.([]any)
	if !isArray {
		return properties, resultOk
	}

	for _, value := range valueReceivedAsArray {
		if property, isPropertyMap := value.(map[string]any); isPropertyMap && property != nil {
			resultOk = true
			break
		}
	}

	if itemsProperty, isArray := valueAttributes["items"].(map[string]any); isArray && resultOk {
		if propertyValuesRaw, isPropertyMap := itemsProperty["properties"].(map[string]any); isPropertyMap {
			properties = propertyValuesRaw
		}
	}

	if resultOk && len(properties) == 0 {
		resultOk = false
	}

	return properties, resultOk
}

// unsetAPIValue is a function that returns an unset value for a given type.
func unsetAPIValue(valueType string) any {
	var unsetValue any

	switch valueType {
	case "boolean":
		unsetValue = false
	case "integer":
		unsetValue = 0
	case "number":
		unsetValue = float64(0)
	case "string":
		unsetValue = ""
	case "array":
		unsetValue = []any{}
	}

	return unsetValue
}

// parsePropertiesFromAPI is a function that returns a map of properties parsed from an API response.
func parsePropertiesFromAPI(
	nestedName string,
	responseMapping map[string]any,
	propertyMapping map[string]any,
) (map[string]any, error) {
	propertyMappingCopy := make(map[string]any, len(propertyMapping))

	for key, value := range propertyMapping {
		valueAttributes, isMap := value.(map[string]any)
		if !isMap {
			return nil, fmt.Errorf("%s...%s: property is not a map", nestedName, key)
		}

		_, aivenTypes, err := userconfig.TerraformTypes(userconfig.SlicedString(valueAttributes["type"]))
		if err != nil {
			return nil, err
		}

		if len(aivenTypes) > 1 {
			return nil, fmt.Errorf("%s...%s: multiple types", nestedName, key)
		}

		typeReceived := aivenTypes[0]

		valueReceived, keyExists := responseMapping[key]
		if !keyExists || valueReceived == nil {
			if typeReceived == "object" {
				continue
			}

			valueReceived = unsetAPIValue(typeReceived)
		}

		var valueReceivedParsed any

		switch typeReceived {
		default:
			switch valueReceivedAsArray := valueReceived.(type) {
			default:
				valueReceivedParsed = valueReceived
			case []any:
				var list []any

				if valueNestedProperties, isArray := hasNestedProperties(valueReceived, valueAttributes); isArray {
					for nestedKey, nestedValue := range valueReceivedAsArray {
						nestedValueAlpha, valueIsMap := nestedValue.(map[string]any)
						if !valueIsMap {
							return nil, fmt.Errorf(
								"%s...%s.%d: slice item is not a map", nestedName, key, nestedKey,
							)
						}

						propertiesParsed, err := parsePropertiesFromAPI(
							nestedName, nestedValueAlpha, valueNestedProperties,
						)
						if err != nil {
							return nil, err
						}

						list = append(list, propertiesParsed)
					}
				} else {
					list = append(list, valueReceivedAsArray...)
				}

				var nestedTypes []string

				if itemKey, isArray := valueAttributes["items"].(map[string]any); isArray {
					if oneOfNumericKey, isArrayNumeric := itemKey["one_of"].([]any); isArrayNumeric {
						for _, nestedValue := range oneOfNumericKey {
							if nestedValueAlpha, valueIsMap := nestedValue.(map[string]any); valueIsMap {
								if nestedValueAlphaType, valueIsString :=
									nestedValueAlpha["type"].(string); valueIsString {
									nestedTypes = append(nestedTypes, nestedValueAlphaType)
								}
							}
						}
					} else {
						_, nestedTypes, err = userconfig.TerraformTypes(userconfig.SlicedString(itemKey["type"]))
						if err != nil {
							return nil, err
						}
					}
				}

				if len(nestedTypes) > 1 {
					if len(list) > 0 {
						listFirstSeries := list[0]

						switch listFirstSeries.(type) {
						case bool:
							key = fmt.Sprintf("%s_boolean", key)
						case int:
							key = fmt.Sprintf("%s_integer", key)
						case float64:
							key = fmt.Sprintf("%s_number", key)
						case string:
							key = fmt.Sprintf("%s_string", key)
						case []any:
							key = fmt.Sprintf("%s_array", key)
						case map[string]any:
							key = fmt.Sprintf("%s_object", key)
						default:
							return nil, fmt.Errorf("%s...%s: no suffix for given type", nestedName, key)
						}

						if key == "ip_filter_string" {
							key = "ip_filter"
						}

						if key == "namespaces_string" {
							key = "namespaces"
						}
					} else {
						for _, nestedValue := range nestedTypes {
							trimmedKey := fmt.Sprintf("%s_%s", key, nestedValue)

							if trimmedKey == "ip_filter_string" {
								trimmedKey = "ip_filter"
							}

							if trimmedKey == "namespaces_string" {
								trimmedKey = "namespaces"
							}

							propertyMappingCopy[trimmedKey] = list
						}

						continue
					}
				}

				valueReceivedParsed = list
			}
		case "object":
			valueReceivedAsAlpha, valueIsMap := valueReceived.(map[string]any)
			if !valueIsMap {
				return nil, fmt.Errorf("%s...%s: representation value is not a map", nestedName, key)
			}

			nestedValues, keyExists := valueAttributes["properties"]
			if !keyExists {
				return nil, fmt.Errorf("%s...%s: properties key not found", nestedName, key)
			}

			nestedValuesAsAlpha, valueIsMap := nestedValues.(map[string]any)
			if !valueIsMap {
				return nil, fmt.Errorf("%s...%s: properties value is not a map", nestedName, key)
			}

			propertiesParsed, err := parsePropertiesFromAPI(nestedName, valueReceivedAsAlpha, nestedValuesAsAlpha)
			if err != nil {
				return nil, err
			}

			valueReceivedParsed = []map[string]any{propertiesParsed}
		}

		propertyMappingCopy[userconfig.EncodeKey(key)] = valueReceivedParsed
	}

	return propertyMappingCopy, nil
}

// FromAPI is a function that returns a slice of properties parsed from an API response.
func FromAPI(
	schemaType userconfig.SchemaType,
	nestedName string,
	responseMapping map[string]any,
) ([]map[string]any, error) {
	var propertiesParsed []map[string]any

	if len(responseMapping) == 0 {
		return propertiesParsed, nil
	}

	propertyRequests, _, err := propsReqs(schemaType, nestedName)
	if err != nil {
		return nil, err
	}

	propertyAttributes, err := parsePropertiesFromAPI(nestedName, responseMapping, propertyRequests)
	if err != nil {
		return nil, err
	}

	propertiesParsed = append(propertiesParsed, propertyAttributes)

	return propertiesParsed, nil
}
