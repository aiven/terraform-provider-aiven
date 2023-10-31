package apiconvert

import (
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// propsReqs is a function that returns a map of properties and required properties from a given schema type and node
// name.
func propsReqs(schemaType userconfig.SchemaType, nodeName string) (map[string]any, map[string]struct{}, error) {
	representationMap, err := userconfig.CachedRepresentationMap(schemaType)
	if err != nil {
		return nil, nil, err
	}

	nodeSchema, exists := representationMap[nodeName]
	if !exists {
		return nil, nil, fmt.Errorf("no schema found for %s (type %d)", nodeName, schemaType)
	}

	schemaAsMap, ok := nodeSchema.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("schema %s (type %d) is not a map", nodeName, schemaType)
	}

	properties, exists := schemaAsMap["properties"]
	if !exists {
		return nil, nil, fmt.Errorf("no properties found for %s (type %d)", nodeName, schemaType)
	}

	propertiesAsMap, ok := properties.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("properties of schema %s (type %d) are not a map", nodeName, schemaType)
	}

	requiredProperties := map[string]struct{}{}
	if requiredPropertiesSlice, exists := schemaAsMap["required"].([]any); exists {
		requiredProperties = userconfig.SliceToKeyedMap(requiredPropertiesSlice)
	}

	return propertiesAsMap, requiredProperties, nil
}
