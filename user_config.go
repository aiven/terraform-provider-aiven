// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/hashicorp/terraform/helper/schema"
)

func readUserConfigJSONSchema(name string) map[string]interface{} {
	box := packr.NewBox("./templates")
	data, err := box.Find(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to read %v: %v", name, err))
	}
	var jsonObject interface{}
	err = json.Unmarshal(data, &jsonObject)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal %v: %v", name, err))
	}
	return jsonObject.(map[string]interface{})
}

var userConfigSchemas = map[string]map[string]interface{}{}

// GetUserConfigSchema returns user configuration definition for given resource type
func GetUserConfigSchema(resourceType string) map[string]interface{} {
	if data, ok := userConfigSchemas[resourceType]; ok {
		return data
	}
	var result map[string]interface{}
	switch resourceType {
	case "endpoint":
		result = readUserConfigJSONSchema("integration_endpoints_user_config_schema.json")
	case "integration":
		result = readUserConfigJSONSchema("integrations_user_config_schema.json")
	case "service":
		result = readUserConfigJSONSchema("service_user_config_schema.json")
	default:
		panic("Unknown resourceType " + resourceType)
	}
	userConfigSchemas[resourceType] = result
	return result
}

// GenerateTerraformUserConfigSchema creates Terraform schema definition for user config based
// on user config JSON schema definition.
func GenerateTerraformUserConfigSchema(data map[string]interface{}) map[string]*schema.Schema {
	properties := data["properties"].(map[string]interface{})
	terraformSchema := make(map[string]*schema.Schema)
	for name, definitionRaw := range properties {
		definition := definitionRaw.(map[string]interface{})
		terraformSchema[encodeKeyName(name)] = generateTerraformUserConfigSchema(name, definition)
	}
	return terraformSchema
}

func generateTerraformUserConfigSchema(key string, definition map[string]interface{}) *schema.Schema {
	valueType, _ := getAivenSchemaType(definition["type"])
	sensitive := false
	if strings.Contains(key, "api_key") || strings.Contains(key, "password") {
		sensitive = true
	}
	var diffFunction schema.SchemaDiffSuppressFunc
	if createOnly, ok := definition["createOnly"]; ok && createOnly.(bool) {
		diffFunction = createOnlyDiffSuppressFunc
	} else if valueType == "object" {
		diffFunction = emptyObjectDiffSuppressFunc
	}
	defaultValue := getAivenSchemaDefaultValue(definition)
	title := definition["title"].(string)
	switch valueType {
	case "number":
		return &schema.Schema{
			Default:          defaultValue,
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Optional:         true,
			Sensitive:        sensitive,
			Type:             schema.TypeFloat,
		}
	case "integer":
		_, isFloat := defaultValue.(float64)
		if isFloat {
			defaultValue = int(defaultValue.(float64))
		}
		return &schema.Schema{
			Default:          defaultValue,
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Optional:         true,
			Sensitive:        sensitive,
			Type:             schema.TypeInt,
		}
	case "boolean":
		return &schema.Schema{
			Default:          defaultValue,
			Description:      title,
			DiffSuppressFunc: diffFunction,
			Optional:         true,
			Sensitive:        sensitive,
			Type:             schema.TypeBool,
		}
	case "string":
		return &schema.Schema{
			Default:          defaultValue,
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
		typeString := itemDefinition["type"].(string)
		switch typeString {
		case "number":
			itemType = schema.TypeFloat
		case "integer":
			itemType = schema.TypeInt
		case "boolean":
			itemType = schema.TypeBool
		case "string":
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

func getAivenSchemaType(value interface{}) (string, bool) {
	switch res := value.(type) {
	case string:
		return res, false
	case []interface{}:
		optional := false
		typeString := ""
		for _, typeOrNullRaw := range res {
			typeOrNull := typeOrNullRaw.(string)
			if typeOrNull == "null" {
				optional = true
			} else {
				typeString = typeOrNull
			}
		}
		return typeString, optional
	default:
		panic(fmt.Sprintf("Unexpected user config schema type: %T / %v", value, value))
	}
}

func getAivenSchemaDefaultValue(definition map[string]interface{}) interface{} {
	valueType, _ := getAivenSchemaType(definition["type"])
	defaultValue, ok := definition["default"]
	if !ok && valueType == "number" {
		defaultValue = -1.0
	} else if !ok && valueType == "integer" {
		defaultValue = -1
	} else if (!ok || defaultValue == nil) && valueType == "boolean" {
		defaultValue = false
	} else if (!ok || defaultValue == nil) && valueType == "string" {
		// Terraform has no way of indicating unset values.
		// Convert "<<value not set>>" to unset when handling request
		defaultValue = "<<value not set>>"
	} else if valueType == "array" {
		defaultValue = []interface{}{}
	} else if valueType == "object" {
		defaultValue = map[string]interface{}{}
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

	entrySchema := GetUserConfigSchema(configType)[entryType].(map[string]interface{})
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
		valueType, _ := getAivenSchemaType(schemaDefinition["type"])
		apiValue, ok := apiUserConfig[key]
		key = encodeKeyName(key)
		if !ok || apiValue == nil {
			// To avoid undesired "changes" for values that are not explicitly defined return
			// default values for anything that is not returned in the API response
			terraformConfig[key] = getAivenSchemaDefaultValue(schemaDefinition)
			continue
		}
		switch valueType {
		case "object":
			res := convertAPIUserConfigToTerraformCompatibleFormat(
				apiValue.(map[string]interface{}), schemaDefinition["properties"].(map[string]interface{}),
			)
			terraformConfig[key] = []map[string]interface{}{res}
		case "integer":
			switch res := apiValue.(type) {
			case float64:
				terraformConfig[key] = int(res)
			case float32:
				terraformConfig[key] = int(res)
			case int64:
				terraformConfig[key] = int(res)
			case int32:
				terraformConfig[key] = int(res)
			default:
				panic(fmt.Sprintf("Unexpected value type for '%v': %v / %T", key, apiValue, apiValue))
			}
		case "number":
			switch res := apiValue.(type) {
			case float64:
				terraformConfig[key] = res
			case float32:
				terraformConfig[key] = float64(res)
			case int64:
				terraformConfig[key] = float64(res)
			case int32:
				terraformConfig[key] = float64(res)
			default:
				panic(fmt.Sprintf("Unexpected value type for '%v': %v / %T", key, apiValue, apiValue))
			}
		default:
			terraformConfig[key] = apiValue
		}
	}
	return terraformConfig
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
	entrySchema := GetUserConfigSchema(configType)[entryType].(map[string]interface{})
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
	convertedValue := value
	omit := false
	valueType, _ := getAivenSchemaType(definition["type"])
	switch valueType {
	case "integer":
		switch res := value.(type) {
		case int:
		case int64:
			convertedValue = int(res)
		case int32:
			convertedValue = int(res)
		default:
			panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected integer", serviceType, value, key))
		}
		minimum, hasMin := definition["minimum"]
		if hasMin && value.(int) < int(minimum.(float64)) {
			omit = true
		}
	case "number":
		switch res := value.(type) {
		case float64:
		case float32:
			convertedValue = float64(res)
		case int64:
			convertedValue = float64(res)
		case int32:
			convertedValue = float64(res)
		case int:
			convertedValue = float64(res)
		default:
			panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected float", serviceType, value, key))
		}
		minimum, hasMin := definition["minimum"]
		if hasMin && value.(float64) < minimum.(float64) {
			omit = true
		}
	case "boolean":
	case "string":
		switch value.(type) {
		case string:
			if value == "<<value not set>>" {
				omit = true
			}
		default:
			panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected string", serviceType, value, key))
		}
	case "object":
		if value == nil {
			omit = true
		} else {
			if asList, isList := value.([]interface{}); isList {
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
			} else if asMap, isMap := value.(map[string]interface{}); isMap {
				convertedValue = convertTerraformUserConfigToAPICompatibleFormat(
					serviceType, newResource, asMap, definition["properties"].(map[string]interface{}),
				)
			} else {
				panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected map", serviceType, value, key))
			}
		}
	case "array":
		if value == nil {
			omit = true
		} else {
			switch value.(type) {
			case []interface{}:
			default:
				panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected list", serviceType, value, key))
			}
			asArray := value.([]interface{})
			if len(asArray) == 0 {
				omit = true
			} else {
				values := make([]interface{}, len(value.([]interface{})))
				itemDefinition := definition["items"].(map[string]interface{})
				for idx, arrValue := range asArray {
					arrValueConverted, _ := convertTerraformUserConfigValueToAPICompatibleFormat(
						serviceType, newResource, key, arrValue, itemDefinition)
					values[idx] = arrValueConverted
				}
				convertedValue = values
			}
		}
	default:
		panic(fmt.Sprintf("Unsupported value type %v for %v user config key %v", definition["type"], serviceType, key))
	}
	return convertedValue, omit
}

func encodeKeyName(key string) string {
	// Terraform does not accept dots in key names but Aiven API has those at least in PG user config
	return strings.Replace(key, ".", "__dot__", -1)
}

func decodeKeyName(key string) string {
	return strings.Replace(key, "__dot__", ".", -1)
}
