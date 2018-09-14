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
	data, err := box.MustBytes(name)
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

var userConfigSchemas = map[string]map[string]interface{}{
	"endpoint":    readUserConfigJSONSchema("integration_endpoints_user_config_schema.json"),
	"integration": readUserConfigJSONSchema("integrations_user_config_schema.json"),
	"service":     readUserConfigJSONSchema("service_user_config_schema.json"),
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
	defaultValue, ok := definition["default"]
	title := definition["title"].(string)
	switch valueType {
	case "number":
		if !ok {
			defaultValue = -1.0
		}
		return &schema.Schema{
			Default:     defaultValue,
			Description: title,
			Optional:    true,
			Sensitive:   sensitive,
			Type:        schema.TypeFloat,
		}
	case "integer":
		if !ok {
			defaultValue = -1.0
		}
		return &schema.Schema{
			Default:     int(defaultValue.(float64)),
			Description: title,
			Optional:    true,
			Sensitive:   sensitive,
			Type:        schema.TypeInt,
		}
	case "boolean":
		if !ok || defaultValue == nil {
			defaultValue = false
		}
		return &schema.Schema{
			Default:     defaultValue,
			Description: title,
			Optional:    true,
			Sensitive:   sensitive,
			Type:        schema.TypeBool,
		}
	case "string":
		if !ok || defaultValue == nil {
			// Terraform has no way of indicating unset values.
			// Convert "<<value not set>>" to unset when handling request
			defaultValue = "<<value not set>>"
		}
		return &schema.Schema{
			Default:     defaultValue,
			Description: title,
			Optional:    true,
			Sensitive:   sensitive,
			Type:        schema.TypeString,
		}
	case "object":
		return &schema.Schema{
			Description: title,
			Elem:        &schema.Resource{Schema: GenerateTerraformUserConfigSchema(definition)},
			MaxItems:    1,
			Optional:    true,
			Type:        schema.TypeList,
		}
	case "array":
		var itemType schema.ValueType
		typeString := definition["items"].(map[string]interface{})["type"].(string)
		switch typeString {
		case "number":
			itemType = schema.TypeFloat
		case "integer":
			itemType = schema.TypeInt
		case "boolean":
			itemType = schema.TypeBool
		case "string":
			itemType = schema.TypeString
		default:
			panic(fmt.Sprintf("Unexpected user config schema array item type: %T / %v", typeString, typeString))
		}
		return &schema.Schema{
			Description: title,
			Elem:        &schema.Schema{Type: itemType},
			MaxItems:    int(definition["maxItems"].(float64)),
			Optional:    true,
			Sensitive:   sensitive,
			Type:        schema.TypeList,
		}
	default:
		panic(fmt.Sprintf("Unexpected user config schema type: %T / %v", valueType, valueType))
	}
}

func getAivenSchemaType(value interface{}) (string, bool) {
	switch value.(type) {
	case string:
		return value.(string), false
	case []interface{}:
		optional := false
		typeString := ""
		for _, typeOrNullRaw := range value.([]interface{}) {
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

// ConvertAPIUserConfigToTerraformCompatibleFormat converts API response to a format that is
// accepted by Terraform; intermediary lists are added as necessary, default values are provided
// for missing keys and type conversions are performed if necessary.
func ConvertAPIUserConfigToTerraformCompatibleFormat(
	configType string,
	entryType string,
	userConfig map[string]interface{},
) []map[string]interface{} {
	if userConfig == nil || len(userConfig) == 0 {
		return []map[string]interface{}{}
	}

	entrySchema := userConfigSchemas[configType][entryType].(map[string]interface{})
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
			switch valueType {
			case "object":
				terraformConfig[key] = map[string]interface{}{}
			case "array":
				terraformConfig[key] = []interface{}{}
			case "number":
				terraformConfig[key] = -1.0
			case "integer":
				terraformConfig[key] = -1
			case "boolean":
				terraformConfig[key] = false
			case "string":
				terraformConfig[key] = "<<value not set>>"
			}
			continue
		}
		switch valueType {
		case "object":
			res := convertAPIUserConfigToTerraformCompatibleFormat(
				apiValue.(map[string]interface{}), schemaDefinition["properties"].(map[string]interface{}),
			)
			terraformConfig[key] = []map[string]interface{}{res}
		case "integer":
			switch apiValue.(type) {
			case float64:
				terraformConfig[key] = int(apiValue.(float64))
			case float32:
				terraformConfig[key] = int(apiValue.(float32))
			case int64:
				terraformConfig[key] = int(apiValue.(int64))
			case int32:
				terraformConfig[key] = int(apiValue.(int32))
			default:
				panic(fmt.Sprintf("Unexpected value type for '%v': %v / %T", key, apiValue, apiValue))
			}
		case "number":
			switch apiValue.(type) {
			case float64:
				terraformConfig[key] = apiValue.(float64)
			case float32:
				terraformConfig[key] = float64(apiValue.(float32))
			case int64:
				terraformConfig[key] = float64(apiValue.(int64))
			case int32:
				terraformConfig[key] = float64(apiValue.(int32))
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
	d *schema.ResourceData,
) map[string]interface{} {
	mainKey := entryType + "_user_config"
	userConfigsRaw, ok := d.GetOk(mainKey)
	if !ok || userConfigsRaw == nil {
		return nil
	}
	entrySchema := userConfigSchemas[configType][entryType].(map[string]interface{})
	entrySchemaProps := entrySchema["properties"].(map[string]interface{})
	return convertTerraformUserConfigToAPICompatibleFormat(
		entryType, userConfigsRaw.([]interface{})[0].(map[string]interface{}), entrySchemaProps)
}

func convertTerraformUserConfigToAPICompatibleFormat(
	serviceType string,
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
		convertedValue, omit := convertTerraformUserConfigValueToAPICompatibleFormat(serviceType, key, value, definition)
		if !omit {
			apiConfig[key] = convertedValue
		}
	}
	return apiConfig
}

func convertTerraformUserConfigValueToAPICompatibleFormat(
	serviceType string,
	key string,
	value interface{},
	definition map[string]interface{},
) (interface{}, bool) {
	convertedValue := value
	omit := false
	valueType, _ := getAivenSchemaType(definition["type"])
	switch valueType {
	case "integer":
		switch value.(type) {
		case int:
		case int64:
			convertedValue = int(value.(int64))
		case int32:
			convertedValue = int(value.(int32))
		default:
			panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected integer", serviceType, value, key))
		}
		minimum, hasMin := definition["minimum"]
		if hasMin && value.(int) < int(minimum.(float64)) {
			omit = true
		}
	case "number":
		switch value.(type) {
		case float64:
		case float32:
			convertedValue = float64(value.(float32))
		case int64:
			convertedValue = float64(value.(int64))
		case int32:
			convertedValue = float64(value.(int32))
		case int:
			convertedValue = float64(value.(int))
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
			switch value.(type) {
			case []interface{}:
			default:
				panic(fmt.Sprintf("Invalid %v user config key type %T for %v, expected map", serviceType, value, key))
			}
			asMap := value.([]interface{})[0].(map[string]interface{})
			if len(asMap) == 0 {
				omit = true
			} else {
				convertedValue = convertTerraformUserConfigToAPICompatibleFormat(
					serviceType, asMap, definition["properties"].(map[string]interface{}),
				)
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
						serviceType, key, arrValue, itemDefinition)
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
