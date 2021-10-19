// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GenerateTerraformUserConfigSchema creates Terraform schema definition for user config based
// on user config JSON schema definition.
func GenerateTerraformUserConfigSchema(data map[string]interface{}) map[string]*schema.Schema {
	if _, ok := data["properties"]; !ok {
		return map[string]*schema.Schema{}
	}

	properties := data["properties"].(map[string]interface{})
	terraformSchema := make(map[string]*schema.Schema)

	for name, definitionRaw := range properties {
		definition := definitionRaw.(map[string]interface{})
		terraformSchema[encodeKeyName(name)] = generateTerraformUserConfigSchema(name, definition)
	}

	return terraformSchema
}

func generateTerraformUserConfigSchema(key string, definition map[string]interface{}) *schema.Schema {
	valueType := getAivenSchemaType(definition["type"])
	sensitive := false

	if strings.Contains(key, "api_key") || strings.Contains(key, "password") {
		sensitive = true
	}

	var diffFunction schema.SchemaDiffSuppressFunc
	if createOnly, ok := definition["createOnly"]; ok && createOnly.(bool) {
		diffFunction = createOnlyDiffSuppressFunc
	} else if valueType == "object" {
		diffFunction = emptyObjectDiffSuppressFuncSkipArrays(GenerateTerraformUserConfigSchema(definition))
	}

	title := definition["title"].(string)

	switch valueType {
	case "string", "integer", "boolean", "number":
		return &schema.Schema{
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Optional:         true,
			Sensitive:        sensitive,
			Type:             schema.TypeString,
		}
	case "object":
		return &schema.Schema{
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Elem:             &schema.Resource{Schema: GenerateTerraformUserConfigSchema(definition)},
			MaxItems:         1,
			Optional:         true,
			Type:             schema.TypeList,
		}
	case "array":
		var itemType schema.ValueType
		itemDefinition := definition["items"].(map[string]interface{})
		itemDefinition = selectFirstSchemaFromOneOf(itemDefinition)

		typeString := getAivenSchemaType(itemDefinition["type"])
		switch typeString {
		case "string", "integer", "boolean", "number":
			itemType = schema.TypeString
		case "object":
			itemType = schema.TypeList
		default:
			panic(fmt.Sprintf("Unexpected user config schema array item type: %T / %v", typeString, typeString))
		}
		maxItemsVal, maxItemsFound := definition["maxItems"]
		maxItems := 0
		if maxItemsFound {
			maxItems = int(maxItemsVal.(float64))
		}
		var valueDiffFunction schema.SchemaDiffSuppressFunc
		if key == "ip_filter" {
			diffFunction = ipFilterArrayDiffSuppressFunc
			valueDiffFunction = ipFilterValueDiffSuppressFunc
		}
		var elem interface{}
		if itemType == schema.TypeList {
			elem = &schema.Resource{Schema: GenerateTerraformUserConfigSchema(itemDefinition)}
		} else {
			elem = &schema.Schema{
				DiffSuppressFunc: valueDiffFunction,
				Type:             itemType,
			}
		}
		return &schema.Schema{
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Elem:             elem,
			MaxItems:         maxItems,
			Optional:         true,
			Sensitive:        sensitive,
			Type:             schema.TypeList,
		}
	default:
		panic(fmt.Sprintf("Unexpected user config schema type: %T / %v", valueType, valueType))
	}
}

func getAivenSchemaType(value interface{}) string {
	switch res := value.(type) {
	case string:
		return res
	case []interface{}:
		typeString := ""
		for _, typeOrNullRaw := range res {
			typeOrNull := typeOrNullRaw.(string)
			if typeOrNull == "null" {
			} else {
				typeString = typeOrNull
			}
		}
		return typeString
	default:
		panic(fmt.Sprintf("Unexpected user config schema type: %T / %v", value, value))
	}
}

func getAivenSchemaDefaultValue(definition map[string]interface{}) interface{} {
	var defaultValue interface{}

	valueType := getAivenSchemaType(definition["type"])
	switch valueType {
	case "array":
		defaultValue = []interface{}{}
	case "object":
		defaultValue = []map[string]interface{}{}
	default:
		defaultValue = ""
	}

	return defaultValue
}

// ConvertAPIUserConfigToTerraformCompatibleFormat converts API response to a format that is
// accepted by Terraform; intermediary lists are added as necessary, default values are provided
// for missing keys and type conversions are performed if necessary.
func ConvertAPIUserConfigToTerraformCompatibleFormat(
	configType string,
	entryType string,
	userConfig map[string]interface{},
) []map[string]interface{} {
	if len(userConfig) == 0 {
		return []map[string]interface{}{}
	}

	entrySchema := templates.GetUserConfigSchema(configType)[entryType].(map[string]interface{})
	entrySchemaProps := entrySchema["properties"].(map[string]interface{})
	return []map[string]interface{}{convertAPIUserConfigToTerraformCompatibleFormat(userConfig, entrySchemaProps)}
}

func convertAPIUserConfigToTerraformCompatibleFormat(
	apiUserConfig map[string]interface{},
	jsonSchema map[string]interface{},
) map[string]interface{} {
	terraformConfig := make(map[string]interface{})

	for key, schemaDefinitionRaw := range jsonSchema {
		schemaDefinition := schemaDefinitionRaw.(map[string]interface{})

		valueType := getAivenSchemaType(schemaDefinition["type"])

		apiValue, ok := apiUserConfig[key]
		key = encodeKeyName(key)
		if !ok || apiValue == nil {
			// To avoid undesired "changes" for values that are not explicitly defined return
			// default values for anything that is not returned in the API response
			apiValue = getAivenSchemaDefaultValue(schemaDefinition)
			if valueType == "object" {
				continue
			}
		}

		switch valueType {
		case "object":
			res := convertAPIUserConfigToTerraformCompatibleFormat(
				apiValue.(map[string]interface{}), schemaDefinition["properties"].(map[string]interface{}),
			)
			terraformConfig[key] = []map[string]interface{}{res}
		default:
			switch value := apiValue.(type) {
			case string:
				terraformConfig[key] = apiValue
			case bool:
				terraformConfig[key] = strconv.FormatBool(apiValue.(bool))
			case float64:
				terraformConfig[key] = strconv.FormatFloat(apiValue.(float64), 'f', -1, 64)
			case float32:
				terraformConfig[key] = strconv.FormatFloat(apiValue.(float64), 'f', -1, 32)
			case int:
				terraformConfig[key] = strconv.Itoa(apiValue.(int))
			case []interface{}:
				if hasNestedUserConfigurationOptionItems(apiValue, schemaDefinition) {
					var list []interface{}
					for _, v := range apiValue.([]interface{}) {
						res := convertAPIUserConfigToTerraformCompatibleFormat(
							v.(map[string]interface{}), schemaDefinition["items"].(map[string]interface{})["properties"].(map[string]interface{}),
						)
						list = append(list, res)
					}
					terraformConfig[key] = list
				} else {
					var list []interface{}
					for _, v := range apiValue.([]interface{}) {
						list = append(list, fmt.Sprintf("%v", v))
					}
					terraformConfig[key] = list
				}

			default:
				panic(fmt.Sprintf("Invalid user config key type %T for %v", value, key))
			}
		}
	}

	return terraformConfig
}

// hasNestedUserConfigurationOptionItems determines if the user configuration option has nested
// items by definition and base on API value.
func hasNestedUserConfigurationOptionItems(apiValue interface{}, schemaDefinition map[string]interface{}) bool {
	var result bool

	// check if API
	//value has nested an items type of map[string]interface{}
	for _, v := range apiValue.([]interface{}) {
		if b, ok := v.(map[string]interface{}); ok && b != nil {
			// check if schemaDefinition has [items] key
			if _, ok := schemaDefinition["items"]; ok {
				// check if schemaDefinition has [items][properties] key
				if _, ok := schemaDefinition["items"].(map[string]interface{})["properties"]; ok {
					result = true
				}
			}
		}
	}

	return result
}

// ConvertTerraformUserConfigToAPICompatibleFormat converts Terraform user configuration to API compatible
// format; Schema-based Terraform configuration requires using TypeList, which adds one extra layer of lists
// that need to be dropped. Also need to drop dummy "unset" replacement values
func ConvertTerraformUserConfigToAPICompatibleFormat(
	configType string,
	entryType string,
	newResource bool,
	d *schema.ResourceData,
) map[string]interface{} {
	mainKey := entryType + "_user_config"
	userConfigsRaw, ok := d.GetOk(mainKey)
	if !ok || userConfigsRaw == nil {
		return nil
	}
	entrySchema := templates.GetUserConfigSchema(configType)[entryType].(map[string]interface{})
	entrySchemaProps := entrySchema["properties"].(map[string]interface{})
	return convertTerraformUserConfigToAPICompatibleFormat(
		entryType, newResource, userConfigsRaw.([]interface{})[0].(map[string]interface{}), entrySchemaProps)
}

func convertTerraformUserConfigToAPICompatibleFormat(
	serviceType string,
	newResource bool,
	userConfig map[string]interface{},
	configSchema map[string]interface{},
) map[string]interface{} {
	apiConfig := make(map[string]interface{})

	for key, value := range userConfig {
		key = decodeKeyName(key)
		definitionRaw, ok := configSchema[key]
		if !ok {
			panic(fmt.Sprintf("Unsupported %v user config key %v", serviceType, key))
		}
		if definitionRaw == nil {
			continue
		}
		definition := definitionRaw.(map[string]interface{})
		createOnly, ok := definition["createOnly"]
		if ok && createOnly.(bool) && !newResource {
			continue
		}
		convertedValue, omit := convertTerraformUserConfigValueToAPICompatibleFormat(
			serviceType, newResource, key, value, definition)
		if !omit {
			apiConfig[key] = convertedValue
		}
	}

	return apiConfig
}

func convertTerraformUserConfigValueToAPICompatibleFormat(
	serviceType string,
	newResource bool,
	key string,
	value interface{},
	definition map[string]interface{},
) (interface{}, bool) {
	var err error
	var omit bool
	var convertedValue = value

	// get Aiven API value type
	valueType := getAivenSchemaType(definition["type"])

	if canOmit(value, definition) {
		return nil, true
	}

	switch valueType {
	case "integer":
		convertedValue, err = convertTerraformUserConfigValueToAPICompatibleFormatInteger(value)
	case "number":
		convertedValue, err = convertTerraformUserConfigValueToAPICompatibleFormatNumber(value)
	case "boolean":
		convertedValue, err = convertTerraformUserConfigValueToAPICompatibleFormatBoolean(value)
	case "string":
		convertedValue, err = convertTerraformUserConfigValueToAPICompatibleFormatString(value)
	case "object":
		convertedValue, omit, err = convertTerraformUserConfigValueToAPICompatibleFormatObject(
			value, serviceType, newResource, definition)
	case "array":
		convertedValue, omit, err = convertTerraformUserConfigValueToAPICompatibleFormatArray(
			value, serviceType, newResource, key, definition)
	default:
		err = fmt.Errorf("unsupported value type %v for %v user config key %v", definition["type"], serviceType, key)
	}

	if err != nil {
		panic(fmt.Sprintf("unable to convert %v user config key type %T for %v: err %s",
			serviceType, value, key, err))
	}

	return convertedValue, omit
}

// canOmit checks if values can be omitted
func canOmit(value interface{}, definition map[string]interface{}) bool {
	// all empty string values indicate that user configuration option is not by the user
	if value == "" {
		return true
	}

	// for backwards compatibility with the old versions omit when <<value not set>>
	if value == "<<value not set>>" {
		return true
	}

	// if minimum values can be lower then zero do not omit -1
	if minimum, ok := definition["minimum"]; ok {
		if math.Signbit(minimum.(float64)) && value == "-1" {
			return false
		}
	}

	// for backwards compatibility with the old versions omit when -1
	if value == "-1" {
		return true
	}

	return false
}

func convertTerraformUserConfigValueToAPICompatibleFormatArray(value interface{},
	serviceType string,
	newResource bool,
	key string,
	definition map[string]interface{}) (interface{}, bool, error) {
	var convertedValue interface{}
	omit := true

	var empty []interface{}
	if !newResource {
		empty = []interface{}{}
		omit = false
	}

	// when value is nil
	if value == nil {
		return empty, omit, nil
	}

	switch value.(type) {
	case []interface{}:
		asArray := value.([]interface{})

		if len(asArray) == 0 {
			return empty, omit, nil
		}

		values := make([]interface{}, len(value.([]interface{})))
		itemDefinition := definition["items"].(map[string]interface{})
		itemDefinition = selectFirstSchemaFromOneOf(itemDefinition)

		for idx, arrValue := range asArray {
			arrValueConverted, _ := convertTerraformUserConfigValueToAPICompatibleFormat(
				serviceType, newResource, key, arrValue, itemDefinition)
			values[idx] = arrValueConverted
		}

		convertedValue = values
		omit = false
	default:
		return nil, false, fmt.Errorf("invalid %v user config key type %T for %v, expected list", serviceType, value, key)
	}

	return convertedValue, omit, nil
}

func selectFirstSchemaFromOneOf(itemDefinition map[string]interface{}) map[string]interface{} {
	if oneOf, ok := itemDefinition["oneOf"]; ok {
		if types, ok := oneOf.([]interface{}); ok && len(types) > 0 {
			itemDefinition = types[0].(map[string]interface{})
		}
	}
	return itemDefinition
}

func convertTerraformUserConfigValueToAPICompatibleFormatObject(
	value interface{},
	serviceType string,
	newResource bool,
	definition map[string]interface{}) (interface{}, bool, error) {
	var convertedValue interface{}

	// when value is nil
	if value == nil {
		return nil, true, nil
	}

	// when value is TypeList
	if asList, isList := value.([]interface{}); isList {
		var omit bool
		if len(asList) == 0 || (len(asList) == 1 && asList[0] == nil) {
			omit = true
		} else {
			asMap := asList[0].(map[string]interface{})
			if len(asMap) == 0 {
				omit = true
			} else {
				convertedValue = convertTerraformUserConfigToAPICompatibleFormat(
					serviceType, newResource, asMap, definition["properties"].(map[string]interface{}),
				)
			}
		}

		return convertedValue, omit, nil
	}

	// when value is TypeMap
	if asMap, isMap := value.(map[string]interface{}); isMap {
		convertedValue = convertTerraformUserConfigToAPICompatibleFormat(
			serviceType, newResource, asMap, definition["properties"].(map[string]interface{}),
		)

		return convertedValue, false, nil
	}

	return nil, false, fmt.Errorf("expected map but got %s", value)
}

func convertTerraformUserConfigValueToAPICompatibleFormatInteger(value interface{}) (int, error) {
	var convertedValue int

	switch value := value.(type) {
	case int:
		convertedValue = value
	case string:
		var err error
		convertedValue, err = strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("impossible to convert int to a string: %s", err)
		}
	default:
		return 0, fmt.Errorf("expected int or string but got %s", value)
	}

	return convertedValue, nil
}

func convertTerraformUserConfigValueToAPICompatibleFormatNumber(value interface{}) (float64, error) {
	var convertedValue float64

	switch res := value.(type) {
	case float64:
		convertedValue = res
	case string:
		var err error
		convertedValue, err = strconv.ParseFloat(value.(string), 64)
		if err != nil {
			return 0, fmt.Errorf("impossible to convert float64 to a string: %s", err)
		}
	default:
		return 0, fmt.Errorf("expected float64 or string but got %s", value)
	}

	return convertedValue, nil
}

func convertTerraformUserConfigValueToAPICompatibleFormatBoolean(value interface{}) (bool, error) {
	var convertedValue bool

	switch value := value.(type) {
	case string:
		var err error
		convertedValue, err = strconv.ParseBool(value)
		if err != nil {
			return false, err
		}
	case bool:
		convertedValue = value
	default:
		return false, fmt.Errorf("expected boolean or string but got %s", value)
	}

	return convertedValue, nil
}

func convertTerraformUserConfigValueToAPICompatibleFormatString(value interface{}) (string, error) {
	var convertedValue string

	switch value := value.(type) {
	case string:
		convertedValue = value
	default:
		return "", fmt.Errorf("expected string but got %s", value)
	}

	return convertedValue, nil
}

func encodeKeyName(key string) string {
	// Terraform does not accept dots in key names but Aiven API has those at least in PG user config
	return strings.Replace(key, ".", "__dot__", -1)
}

func decodeKeyName(key string) string {
	return strings.Replace(key, "__dot__", ".", -1)
}
