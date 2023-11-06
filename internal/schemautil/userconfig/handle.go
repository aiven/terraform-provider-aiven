//nolint:unused
package userconfig

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"golang.org/x/exp/maps"
)

// handlePrimitiveTypeProperty is a function that converts a primitive type property to a Terraform schema.
func handlePrimitiveTypeProperty(
	name string,
	property map[string]any,
	typeStr string,
	isRequired bool,
) map[string]*jen.Statement {
	return map[string]*jen.Statement{
		name: jen.Values(convertPropertyToSchema(name, property, typeStr, true, isRequired)),
	}
}

// handleObjectProperty converts an object type property to a Terraform schema.
func handleObjectProperty(
	objectName string,
	propertyMap map[string]any,
	typeString string,
	requiredProperties map[string]struct{},
) (map[string]*jen.Statement, error) {
	properties, propertiesExist := propertyMap["properties"].(map[string]any)
	if !propertiesExist {
		itemsAt, itemsExist := propertyMap["items"].(map[string]any)
		if itemsExist {
			properties, propertiesExist = itemsAt["properties"].(map[string]any)
		}

		if !propertiesExist {
			return nil, fmt.Errorf("unable to get properties field: %#v", propertyMap)
		}
	}

	resourceStatements := convertPropertyToSchema(
		objectName, propertyMap, typeString, true, false,
	)

	schemaMapAt, err := convertPropertiesToSchemaMap(properties, requiredProperties)
	if err != nil {
		return nil, err
	}

	schemaValues := jen.Map(jen.String()).Op("*").Qual(SchemaPackage, "Schema").Values(schemaMapAt)

	resourceStatements[jen.Id("Elem")] = jen.Op("&").Qual(SchemaPackage, "Resource").Values(jen.Dict{
		jen.Id("Schema"): schemaValues,
	})

	// TODO: Check if we can access the schema via diff suppression function.
	resourceStatements[jen.Id("DiffSuppressFunc")] = jen.Qual(
		SchemaUtilPackage, "EmptyObjectDiffSuppressFuncSkipArrays",
	).Call(schemaValues)

	resourceStatements[jen.Id("MaxItems")] = jen.Lit(1)

	return map[string]*jen.Statement{objectName: jen.Values(resourceStatements)}, nil
}

// handleArrayOfPrimitiveTypeProperty is a function that converts an array of primitive type property to a Terraform
// schema.
func handleArrayOfPrimitiveTypeProperty(propertyName string, terraformType string) *jen.Statement {
	propertyAttributes := jen.Dict{
		jen.Id("Type"): jen.Qual(SchemaPackage, terraformType),
	}

	if propertyName == "ip_filter" {
		// TODO: Add ip_filter_object to this sanity check when DiffSuppressFunc is implemented for it.
		propertyAttributes[jen.Id("DiffSuppressFunc")] = jen.Qual(
			SchemaUtilPackage, "IPFilterValueDiffSuppressFunc",
		)
	}

	return jen.Op("&").Qual(SchemaPackage, "Schema").Values(propertyAttributes)
}

// handleArrayOfAggregateTypeProperty is a function that converts an array of aggregate type property to a Terraform
// schema.
func handleArrayOfAggregateTypeProperty(ip map[string]any, req map[string]struct{}) (*jen.Statement, error) {
	pc, err := convertPropertiesToSchemaMap(ip, req)
	if err != nil {
		return nil, err
	}

	return jen.Op("&").Qual(SchemaPackage, "Resource").Values(jen.Dict{
		jen.Id("Schema"): jen.Map(jen.String()).Op("*").Qual(SchemaPackage, "Schema").Values(pc),
	}), nil
}

// handleArrayProperty is a function that converts an array type property to a Terraform schema.
func handleArrayProperty(
	propertyName string,
	propertyMap map[string]any,
	terraformType string,
) (map[string]*jen.Statement, error) {
	itemAttributes, ok := propertyMap["items"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("items is not a map[string]any: %#v", propertyMap)
	}

	var element *jen.Statement

	var terraformNames, aivenTypeNames []string

	var err error

	oneOfOptions, isOneOf := itemAttributes["one_of"].([]any)
	if isOneOf {
		var complexTypes []string

		for _, v := range oneOfOptions {
			oneOfMap, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("one_of element is not a map[string]any: %#v", v)
			}

			complexTypes = append(complexTypes, oneOfMap["type"].(string))
		}

		terraformNames, aivenTypeNames, err = TerraformTypes(complexTypes)
		if err != nil {
			return nil, err
		}
	} else {
		terraformNames, aivenTypeNames, err = TerraformTypes(SlicedString(itemAttributes["type"]))
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string]*jen.Statement)

	for k, terraformName := range terraformNames {
		adjustedName := propertyName

		if len(terraformNames) > 1 {
			adjustedName = fmt.Sprintf("%s_%s", propertyName, aivenTypeNames[k])

			// TODO: Remove with the next major version.
			if adjustedName == "ip_filter_string" {
				adjustedName = "ip_filter"
			}

			// TODO: Remove with the next major version.
			if adjustedName == "namespaces_string" {
				adjustedName = "namespaces"
			}
		}

		var oneOfItemAttributes map[string]any

		if isOneOf {
			oneOfItemAttributes, ok = oneOfOptions[k].(map[string]any)
			if !ok {
				return nil,
					fmt.Errorf("unable to convert one_of item to map[string]any: %#v", oneOfOptions[k])
			}
		}

		if isTerraformTypePrimitive(terraformName) {
			element = handleArrayOfPrimitiveTypeProperty(adjustedName, terraformName)
		} else {
			var itemProperties map[string]any

			if isOneOf {
				itemProperties, ok = oneOfItemAttributes["properties"].(map[string]any)
				if !ok {
					return nil, fmt.Errorf(
						"unable to convert one_of item properties to map[string]any: %#v",
						oneOfItemAttributes,
					)
				}
			} else {
				itemProperties, ok = itemAttributes["properties"].(map[string]any)
				if !ok {
					return nil,
						fmt.Errorf("could not find properties in an array of aggregate type: %#v", propertyMap)
				}
			}

			requiredProperties := map[string]struct{}{}

			if requiredItems, ok := itemAttributes["required"].([]any); ok {
				requiredProperties = SliceToKeyedMap(requiredItems)
			}

			element, err = handleArrayOfAggregateTypeProperty(itemProperties, requiredProperties)
			if err != nil {
				return nil, err
			}
		}

		schema := convertPropertyToSchema(propertyName, propertyMap, terraformType, !isOneOf, false)

		if isOneOf {
			oneOfType, ok := oneOfItemAttributes["type"].(string)
			if !ok {
				return nil, fmt.Errorf("one_of item type is not a string: %#v", oneOfItemAttributes)
			}

			_, defaultPropertyDescription := descriptionForProperty(propertyMap, terraformType)

			deprecationIndicator, oneOfItemDescription := descriptionForProperty(oneOfItemAttributes, oneOfType)

			schema[jen.Id("Description")] = jen.Lit(
				fmt.Sprintf("%s %s", defaultPropertyDescription, oneOfItemDescription),
			)

			if deprecationIndicator {
				schema[jen.Id("Deprecated")] = jen.Lit("Usage of this field is discouraged.")
			}
		}

		schema[jen.Id("Elem")] = element

		if adjustedName == "ip_filter" {
			// TODO: Add ip_filter_object to this sanity check when DiffSuppressFunc is implemented for it.
			schema[jen.Id("DiffSuppressFunc")] = jen.Qual(
				SchemaUtilPackage, "IPFilterArrayDiffSuppressFunc",
			)
		}

		if maxItems, ok := propertyMap["max_items"].(int); ok {
			schema[jen.Id("MaxItems")] = jen.Lit(maxItems)
		}

		orderedSchema := jen.Dict{}
		for key, value := range schema {
			orderedSchema[key] = value
		}

		// TODO: Remove with the next major version.
		if adjustedName == "ip_filter" || (isOneOf && adjustedName == "namespaces") {
			schema[jen.Id("Deprecated")] = jen.Lit(
				fmt.Sprintf("This will be removed in v5.0.0 and replaced with %s_string instead.", adjustedName),
			)
		}

		result[adjustedName] = jen.Values(schema)

		if adjustedName == "ip_filter" || (isOneOf && adjustedName == "namespaces") {
			result[fmt.Sprintf("%s_string", adjustedName)] = jen.Values(orderedSchema)
		}
	}

	return result, nil
}

// handleAggregateTypeProperty is a function that converts an aggregate type property to a Terraform schema.
func handleAggregateTypeProperty(
	propertyName string,
	propertyAttributes map[string]any,
	terraformType string,
	aivenType string,
) (map[string]*jen.Statement, error) {
	resultStatements := make(map[string]*jen.Statement)

	requiredProperties := map[string]struct{}{}

	if requiredSlice, ok := propertyAttributes["required"].([]any); ok {
		requiredProperties = SliceToKeyedMap(requiredSlice)
	}

	switch aivenType {
	case "object":
		objectStatements, err := handleObjectProperty(
			propertyName, propertyAttributes, terraformType, requiredProperties,
		)
		if err != nil {
			return nil, err
		}

		maps.Copy(resultStatements, objectStatements)
	case "array":
		arrayStatements, err := handleArrayProperty(propertyName, propertyAttributes, terraformType)
		if err != nil {
			return nil, err
		}

		maps.Copy(resultStatements, arrayStatements)
	default:
		return nil, fmt.Errorf("unknown aggregate type: %s", aivenType)
	}

	return resultStatements, nil
}
