//nolint:unused
package userconfig

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/aiven/go-api-schemas/pkg/dist"
	"gopkg.in/yaml.v3"
)

// SchemaType represents a custom type for Terraform schema.
type SchemaType int

const (
	// ServiceTypes represents the service schema type.
	ServiceTypes SchemaType = iota

	// IntegrationTypes represents the integration schema type.
	IntegrationTypes

	// IntegrationEndpointTypes represents the integration endpoint schema type.
	IntegrationEndpointTypes
)

var (
	// cachedRepresentationMaps is a map of cached representation maps.
	cachedRepresentationMaps = make(map[SchemaType]map[string]any, 3)

	// cachedRepresentationMapsMutex is a mutex for the cached representation maps.
	cachedRepresentationMapsMutex = sync.Mutex{}

	// typeSuffixRegExp is a regular expression that matches type suffixes.
	typeSuffixRegExp = regexp.MustCompile(`^.*_(boolean|integer|number|string|array|object)$`)
)

// CachedRepresentationMap returns a cached representation map for a given schema type.
func CachedRepresentationMap(schemaType SchemaType) (map[string]any, error) {
	if _, ok := map[SchemaType]struct{}{
		ServiceTypes:             {},
		IntegrationTypes:         {},
		IntegrationEndpointTypes: {},
	}[schemaType]; !ok {
		return nil, fmt.Errorf("unknown schema type: %d", schemaType)
	}

	switch schemaType {
	case ServiceTypes:
		return representationToMap(schemaType, dist.ServiceTypes)
	case IntegrationTypes:
		return representationToMap(schemaType, dist.IntegrationTypes)
	case IntegrationEndpointTypes:
		return representationToMap(schemaType, dist.IntegrationEndpointTypes)
	default:
		return nil, fmt.Errorf("unknown schema type %d", schemaType)
	}
}

// representationToMap converts a YAML representation of a Terraform schema to a map.
func representationToMap(schemaType SchemaType, representation []byte) (map[string]any, error) {
	cachedRepresentationMapsMutex.Lock()
	defer cachedRepresentationMapsMutex.Unlock()

	if cachedMap, ok := cachedRepresentationMaps[schemaType]; ok {
		return cachedMap, nil
	}

	var mapRepresentation map[string]any
	if err := yaml.Unmarshal(representation, &mapRepresentation); err != nil {
		return nil, err
	}

	cachedRepresentationMaps[schemaType] = mapRepresentation
	return mapRepresentation, nil
}

// TerraformTypes converts schema representation types to Terraform types.
func TerraformTypes(types []string) ([]string, []string, error) {
	var terraformTypes, aivenTypes []string

	for _, typeValue := range types {
		switch typeValue {
		case "null":
			// TODO: Handle this case.
			// This is a special case where the value can be null.
			// There should be a default value set for this case.
			continue
		case "boolean":
			terraformTypes = append(terraformTypes, "TypeBool")
		case "integer":
			terraformTypes = append(terraformTypes, "TypeInt")
		case "number":
			terraformTypes = append(terraformTypes, "TypeFloat")
		case "string":
			terraformTypes = append(terraformTypes, "TypeString")
		case "array", "object":
			terraformTypes = append(terraformTypes, "TypeList")
		default:
			return nil, nil, fmt.Errorf("unknown type: %s", typeValue)
		}

		aivenTypes = append(aivenTypes, typeValue)
	}

	return terraformTypes, aivenTypes, nil
}

// isTerraformTypePrimitive checks if a Terraform type is a primitive type.
func isTerraformTypePrimitive(terraformType string) bool {
	switch terraformType {
	case "TypeBool", "TypeInt", "TypeFloat", "TypeString":
		return true
	default:
		return false
	}
}

// mustStringSlice converts an interface to a slice of strings.
func mustStringSlice(value any) ([]string, error) {
	valueAsSlice, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("not a slice: %#v", value)
	}

	stringSlice := make([]string, len(valueAsSlice))

	for index, value := range valueAsSlice {
		stringValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("value is not a string: %#v", value)
		}

		stringSlice[index] = stringValue
	}

	return stringSlice, nil
}

// SlicedString accepts a string or a slice of strings and returns a slice of strings.
func SlicedString(value any) []string {
	valueAsSlice, ok := value.([]any)
	if ok {
		stringSlice, err := mustStringSlice(valueAsSlice)
		if err != nil {
			panic(err)
		}

		return stringSlice
	}

	valueAsString, ok := value.(string)
	if !ok {
		panic(fmt.Sprintf("value is not a string or a slice of strings: %#v", value))
	}

	return []string{valueAsString}
}

// constDescriptionReplaceables is a map of strings that are replaced in descriptions.
var constDescriptionReplaceables = map[string]string{
	"DEPRECATED: ":                 "",
	"This setting is deprecated. ": "",
	"[seconds]":                    "(seconds)",
}

// descriptionForProperty returns the description for a property.
func descriptionForProperty(
	property map[string]any,
	terraformType string,
) (isDeprecated bool, description string) {
	if descriptionValue, ok := property["description"].(string); ok {
		description = descriptionValue
	} else {
		description = property["title"].(string)
	}

	isDeprecated = strings.Contains(strings.ToLower(description), "deprecated")

	// shouldCapitalize is a flag indicating if the first letter should be capitalized.
	shouldCapitalize := false

	// Some descriptions have a built-in deprecation notice, so we need to remove it.
	for old, new := range constDescriptionReplaceables {
		previousDescription := description

		description = strings.ReplaceAll(description, old, new)

		if previousDescription != description {
			shouldCapitalize = true
		}
	}

	descriptionBuilder := Desc(description)

	if shouldCapitalize {
		descriptionBuilder = descriptionBuilder.ForceFirstLetterCapitalization()
	}

	if defaultValue, ok := property["default"]; ok && isTerraformTypePrimitive(terraformType) {
		skipDefaultValue := false

		if defaultValueAsString, ok := defaultValue.(string); ok {
			if defaultValueAsString == "" {
				skipDefaultValue = true
			}
		}

		if !skipDefaultValue {
			descriptionBuilder = descriptionBuilder.DefaultValue(defaultValue)
		}
	}

	description = descriptionBuilder.Build()

	return isDeprecated, description
}

// EncodeKey encodes a key for a Terraform schema.
func EncodeKey(key string) string {
	return strings.ReplaceAll(key, ".", "__dot__")
}

// DecodeKey decodes a key for a Terraform schema.
func DecodeKey(key string) string {
	return strings.ReplaceAll(key, "__dot__", ".")
}

// IsKeyTyped checks if a key is typed, i.e., has a type suffix in it.
func IsKeyTyped(key string) bool {
	return typeSuffixRegExp.MatchString(key)
}

// SliceToKeyedMap converts a slice of any type to a map with keys of type string.
// It expects that all elements in the slice are of type string, otherwise it will panic.
// The values in the map are of type struct{} to minimize memory usage, as we are only interested in the keys.
func SliceToKeyedMap(slice []any) map[string]struct{} {
	// Initialize an empty map with string keys and struct{} values.
	result := make(map[string]struct{})

	// Iterate through each element in the slice.
	for _, value := range slice {
		// Assert that the element is of type string, and then use it as a key in the map.
		// The value associated with each key is an empty struct{}.
		result[value.(string)] = struct{}{}
	}

	// Return the resulting map.
	return result
}
