//nolint:unused
package userconfig

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

// convertPropertyToSchema is a function that converts a property to a Terraform schema.
func convertPropertyToSchema(
	propertyName string,
	propertyAttributes map[string]any,
	terraformType string,
	addDescription bool,
	isRequired bool,
) jen.Dict {
	resultDict := jen.Dict{
		jen.Id("Type"): jen.Qual(SchemaPackage, terraformType),
	}

	if addDescription {
		isDeprecated, description := descriptionForProperty(propertyAttributes, terraformType)

		resultDict[jen.Id("Description")] = jen.Lit(description)

		if isDeprecated {
			resultDict[jen.Id("Deprecated")] = jen.Lit("Usage of this field is discouraged.")
		}
	}

	if isRequired {
		resultDict[jen.Id("Required")] = jen.Lit(true)
	} else {
		resultDict[jen.Id("Optional")] = jen.Lit(true)

		if defaultValue, ok := propertyAttributes["default"]; ok && isTerraformTypePrimitive(terraformType) {
			resultDict[jen.Id("Default")] = jen.Lit(defaultValue)
		}
	}

	if createOnly, ok := propertyAttributes["create_only"]; ok && createOnly.(bool) {
		resultDict[jen.Id("ForceNew")] = jen.Lit(true)
	}

	if strings.Contains(propertyName, "api_key") || strings.Contains(propertyName, "password") {
		resultDict[jen.Id("Sensitive")] = jen.Lit(true)
	}

	// TODO: Generate validation rules for generated schema properties, also validate that value is within enum values.

	return resultDict
}

// convertPropertiesToSchemaMap is a function that converts a map of properties to a map of Terraform schemas.
func convertPropertiesToSchemaMap(properties map[string]any, requiredProperties map[string]struct{}) (jen.Dict, error) {
	resultDict := make(jen.Dict, len(properties))

	for propertyName, propertyValue := range properties {
		propertyAttributes, ok := propertyValue.(map[string]any)
		if !ok {
			continue
		}

		terraformTypes, aivenTypes, err := TerraformTypes(SlicedString(propertyAttributes["type"]))
		if err != nil {
			return nil, err
		}

		if len(terraformTypes) > 1 {
			return nil, fmt.Errorf("multiple types for %s", propertyName)
		}

		terraformType, aivenType := terraformTypes[0], aivenTypes[0]

		_, isRequired := requiredProperties[propertyName]

		var schemaStatements map[string]*jen.Statement

		if isTerraformTypePrimitive(terraformType) {
			schemaStatements = handlePrimitiveTypeProperty(propertyName, propertyAttributes, terraformType, isRequired)
		} else {
			schemaStatements, err = handleAggregateTypeProperty(
				propertyName, propertyAttributes, terraformType, aivenType,
			)
			if err != nil {
				return nil, err
			}
		}

		if schemaStatements == nil {
			continue
		}

		for keyName, valueNode := range schemaStatements {
			resultDict[jen.Lit(EncodeKey(keyName))] = valueNode
		}
	}

	return resultDict, nil
}
