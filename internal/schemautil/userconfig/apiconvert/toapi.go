package apiconvert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// resourceDatable is an interface that allows to get the resource data from the schema.
// This is needed to be able to test the conversion functions. See schema.ResourceData for more.
type resourceDatable interface {
	GetOk(string) (any, bool)
	HasChange(string) bool
	IsNewResource() bool
}

var (
	// keyPathEndingInNumberRegExp is a regular expression that matches a string that matches:
	//   1. key.1.key2.0.key3.2.key.5
	//   2. key123.0
	//   3. key.1
	//   4. key2.9
	//   5. key..8
	// and does not match:
	//   1. key.key2
	//   2. key.01
	//   3. key.abc
	//   4. .1
	//   5. key.
	keyPathEndingInNumberRegExp = regexp.MustCompile(`.+\.[0-9]$`)

	// dotSeparatedNumberRegExp is a regular expression that matches a string that matches:
	//   1. .5 (match: .5)
	//   2. .9. (match: .9.)
	//   3. 0.1 (match: .1)
	//   4. key.2 (match: .2)
	//   5. 1.2.3 (match: .2.)
	//   6. key..8 (match: .8)
	// and does not match:
	//   1. .123
	//   2. 1.
	//   3. 1..
	//   4. .5a
	dotSeparatedNumberRegExp = regexp.MustCompile(`\.\d($|\.)`)
)

// arrayItemToAPI is a function that converts array property of Terraform user configuration schema to API
// compatible format.
func arrayItemToAPI(
	serviceName string,
	fullKeyPath []string,
	arrayKey string,
	arrayValues []any,
	itemMap map[string]any,
	resourceData resourceDatable,
) (any, bool, error) {
	var convertedValues []any

	if len(arrayValues) == 0 {
		return json.RawMessage("[]"), false, nil
	}

	fullKeyString := strings.Join(fullKeyPath, ".")

	// TODO: Remove when this is fixed on backend.
	if arrayKey == "additional_backup_regions" {
		return convertedValues, true, nil
	}

	itemMapItems, ok := itemMap["items"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s (item): items key not found", fullKeyString)
	}

	var itemType string

	// If the key has a type suffix, we use it to determine the type of the value.
	if userconfig.IsKeyTyped(arrayKey) {
		itemType = arrayKey[strings.LastIndexByte(arrayKey, '_')+1:]

		// Find the one_of item that matches the type.
		if oneOfItems, ok := itemMapItems["one_of"]; ok {
			oneOfItemsSlice, ok := oneOfItems.([]any)
			if !ok {
				return nil, false, fmt.Errorf("%s (items.one_of): not a slice", fullKeyString)
			}

			for i, oneOfItem := range oneOfItemsSlice {
				oneOfItemMap, ok := oneOfItem.(map[string]any)
				if !ok {
					return nil, false, fmt.Errorf("%s (items.one_of.%d): not a map", fullKeyString, i)
				}

				if itemTypeValue, ok := oneOfItemMap["type"]; ok && itemTypeValue == itemType {
					itemMapItems = oneOfItemMap

					break
				}
			}
		}
	} else {
		// TODO: Remove this statement and the branch below it with the next major version.
		_, ok := itemMapItems["one_of"]

		if arrayKey == "ip_filter" || (ok && arrayKey == "namespaces") {
			itemType = "string"
		} else {
			_, itemTypes, err := userconfig.TerraformTypes(userconfig.SlicedString(itemMapItems["type"]))
			if err != nil {
				return nil, false, err
			}

			if len(itemTypes) > 1 {
				return nil, false, fmt.Errorf("%s (type): multiple types", fullKeyString)
			}

			itemType = itemTypes[0]
		}
	}

	for i, arrayValue := range arrayValues {
		// We only accept slices there, so we need to nest the value into a slice if the value is of object type.
		if itemType == "object" {
			arrayValue = []any{arrayValue}
		}

		convertedValue, omit, err := itemToAPI(
			serviceName,
			itemType,
			append(fullKeyPath, fmt.Sprintf("%d", i)),
			fmt.Sprintf("%s.%d", arrayKey, i),
			arrayValue,
			itemMapItems,
			false,
			resourceData,
		)
		if err != nil {
			return nil, false, err
		}

		if !omit {
			convertedValues = append(convertedValues, convertedValue)
		}
	}

	return convertedValues, false, nil
}

// objectItemToAPI is a function that converts object property of Terraform user configuration schema to API
// compatible format.
func objectItemToAPI(
	serviceName string,
	fullKeyPath []string,
	objectValues []any,
	itemSchema map[string]any,
	resourceData resourceDatable,
) (any, bool, error) {
	var result any

	fullKeyString := strings.Join(fullKeyPath, ".")

	firstValue := objectValues[0]

	// Object with only "null" fields becomes nil
	// Which can't be cast into a map
	if firstValue == nil {
		return result, true, nil
	}

	firstValueAsMap, ok := firstValue.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s: not a map", fullKeyString)
	}

	itemProperties, ok := itemSchema["properties"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s (item): properties key not found", fullKeyString)
	}

	requiredFields := map[string]struct{}{}

	if schemaRequiredFields, ok := itemSchema["required"].([]any); ok {
		requiredFields = userconfig.SliceToKeyedMap(schemaRequiredFields)
	}

	if !keyPathEndingInNumberRegExp.MatchString(fullKeyString) {
		fullKeyPath = append(fullKeyPath, "0")
	}

	result, err := propsToAPI(
		serviceName,
		fullKeyPath,
		firstValueAsMap,
		itemProperties,
		requiredFields,
		resourceData,
	)
	if err != nil {
		return nil, false, err
	}

	return result, false, nil
}

// itemToAPI is a function that converts property of Terraform user configuration schema to API compatible format.
func itemToAPI(
	serviceName string,
	itemType string,
	fullKeyPath []string,
	key string,
	value any,
	inputMap map[string]any,
	isRequired bool,
	resourceData resourceDatable,
) (any, bool, error) {
	result := value

	fullKeyString := strings.Join(fullKeyPath, ".")

	omitValue := !resourceData.HasChange(fullKeyString)

	if omitValue && len(fullKeyPath) > 3 {
		lastDotWithNumberIndex := dotSeparatedNumberRegExp.FindAllStringIndex(fullKeyString, -1)
		if lastDotWithNumberIndex != nil {
			_, exists := resourceData.GetOk(fullKeyString)
			lengthOfMatches := len(lastDotWithNumberIndex)

			if (exists || !reflect.ValueOf(value).IsZero()) &&
				resourceData.HasChange(fullKeyString[:lastDotWithNumberIndex[lengthOfMatches-(lengthOfMatches-1)][0]]) {
				omitValue = false
			}
		}
	}

	if omitValue && isRequired {
		omitValue = false
	}

	switch itemType {
	case "boolean":
		if _, ok := value.(bool); !ok {
			return nil, false, fmt.Errorf("%s: not a boolean", fullKeyString)
		}
	case "integer":
		if _, ok := value.(int); !ok {
			return nil, false, fmt.Errorf("%s: not an integer", fullKeyString)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return nil, false, fmt.Errorf("%s: not a number", fullKeyString)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return nil, false, fmt.Errorf("%s: not a string", fullKeyString)
		}
	case "array", "object":
		valueArray, ok := value.([]any)
		if !ok {
			return nil, false, fmt.Errorf("%s: not a slice", fullKeyString)
		}

		if valueArray == nil || omitValue {
			return nil, true, nil
		}

		if itemType == "array" {
			return arrayItemToAPI(serviceName, fullKeyPath, key, valueArray, inputMap, resourceData)
		}

		if len(valueArray) == 0 {
			return nil, true, nil
		}

		return objectItemToAPI(serviceName, fullKeyPath, valueArray, inputMap, resourceData)
	default:
		return nil, false, fmt.Errorf("%s: unsupported type %s", fullKeyString, itemType)
	}

	return result, omitValue, nil
}

// processManyToOneKeys processes many to one keys by mapping them to their first non-empty value.
func processManyToOneKeys(result map[string]any) {
	// manyToOneKeyMap maps primary keys to their associated many-to-one keys.
	manyToOneKeyMap := make(map[string][]string)

	// Iterate over the result map.
	// TODO: Remove all ip_filter and namespaces special cases when these fields are removed.
	for key, value := range result {
		// If the value is a map, process it recursively.
		if valueAsMap, ok := value.(map[string]any); ok {
			processManyToOneKeys(valueAsMap)
		}

		// Ignore keys that are not typed and are not special keys.
		if !userconfig.IsKeyTyped(key) && key != "ip_filter" && key != "namespaces" {
			continue
		}

		// Extract the real key, which is the key without suffix unless it's a special key.
		realKey := key
		if key != "ip_filter" && key != "namespaces" {
			realKey = key[:strings.LastIndexByte(key, '_')]
		}

		// Append the key to its corresponding list in the manyToOneKeyMap map.
		manyToOneKeyMap[realKey] = append(manyToOneKeyMap[realKey], key)
	}

	// By this stage, the 'manyToOneKeyMap' map takes a form similar to the following:
	// map[string][]string{
	//  // For 'ip_filter', there are two associated keys in the user configuration. The first non-empty one is used,
	//  // for instance, if the user shifts from 'ip_filter' to 'ip_filter_object', the latter is preferred.
	// 	"ip_filter": []string{"ip_filter", "ip_filter_object"},
	//  // For 'namespaces', only a single key is present in the user configuration, so it's directly used.
	// 	"namespaces": []string{"namespaces"},
	// }

	// Iterate over the many-to-one keys.
	for primaryKey, associatedKeys := range manyToOneKeyMap {
		var newValue any // The new value for the key.

		wasDeleted := false // Track if any key was deleted in the loop.

		// Attempt to process the values as []any.
		for _, associatedKey := range associatedKeys {
			if associatedValue, ok := result[associatedKey].([]any); ok && len(associatedValue) > 0 {
				newValue = associatedValue

				delete(result, associatedKey) // Delete the processed key-value pair from the result.

				wasDeleted = true
			}
		}

		// If no key was deleted, attempt to process the values as json.RawMessage.
		if !wasDeleted {
			for _, associatedKey := range associatedKeys {
				if associatedValue, ok := result[associatedKey].(json.RawMessage); ok {
					newValue = associatedValue

					delete(result, associatedKey) // Delete the processed key-value pair from the result.

					break
				}
			}
		}

		result[primaryKey] = newValue // Set the new value for the primary key.
	}
}

// propsToAPI is a function that converts properties of Terraform user configuration schema to API compatible format.
func propsToAPI(
	name string,
	fullKeyPath []string,
	types map[string]any,
	properties map[string]any,
	requiredFields map[string]struct{},
	data resourceDatable,
) (map[string]any, error) {
	result := make(map[string]any, len(types))

	fullKeyString := strings.Join(fullKeyPath, ".")

	for typeKey, typeValue := range types {
		typeKey = userconfig.DecodeKey(typeKey)

		rawKey := typeKey

		if userconfig.IsKeyTyped(typeKey) {
			rawKey = typeKey[:strings.LastIndexByte(typeKey, '_')]
		}

		property, ok := properties[rawKey]
		if !ok {
			return nil, fmt.Errorf("%s.%s: key not found", fullKeyString, typeKey)
		}

		if property == nil {
			continue
		}

		propertyAttributes, ok := property.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s.%s: not a map", fullKeyString, typeKey)
		}

		if createOnly, ok := propertyAttributes["create_only"]; ok && createOnly.(bool) && !data.IsNewResource() {
			continue
		}

		_, attributeTypes, err := userconfig.TerraformTypes(userconfig.SlicedString(propertyAttributes["type"]))
		if err != nil {
			return nil, err
		}

		if len(attributeTypes) > 1 {
			return nil, fmt.Errorf("%s.%s.type: multiple types", fullKeyString, typeKey)
		}

		_, isRequired := requiredFields[typeKey]

		attributeType := attributeTypes[0]

		convertedValue, omit, err := itemToAPI(
			name,
			attributeType,
			append(fullKeyPath, typeKey),
			typeKey,
			typeValue,
			propertyAttributes,
			isRequired,
			data,
		)
		if err != nil {
			return nil, err
		}

		if !omit {
			result[typeKey] = convertedValue
		}
	}

	processManyToOneKeys(result)

	return result, nil
}

// ToAPI is a function that converts filled Terraform user configuration schema to API compatible format.
func ToAPI(
	schemaType userconfig.SchemaType,
	serviceName string,
	resourceData resourceDatable,
) (map[string]any, error) {
	var result map[string]any

	fullKeyPath := []string{fmt.Sprintf("%s_user_config", serviceName)}

	terraformConfig, ok := resourceData.GetOk(fullKeyPath[0])
	if !ok || terraformConfig == nil {
		return result, nil
	}

	configSlice, ok := terraformConfig.([]any)
	if !ok {
		return nil, fmt.Errorf("%s (%d): not a slice", serviceName, schemaType)
	}

	firstConfig := configSlice[0]
	if firstConfig == nil {
		return result, nil
	}

	configMap, ok := firstConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s.0 (%d): not a map", serviceName, schemaType)
	}

	properties, requiredProperties, err := propsReqs(schemaType, serviceName)
	if err != nil {
		return nil, err
	}

	result, err = propsToAPI(
		serviceName,
		append(fullKeyPath, "0"),
		configMap,
		properties,
		requiredProperties,
		resourceData,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
