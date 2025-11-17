package util

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// MapModifier modifies request and response objects.
// `r“ is a map used to unmarshal a request or response object.
// `plan` refers to "plan" for Create/Update operations (even when calling Read) or "state" for Read operations.
// The map's keys correspond to the original API keys, not the Terraform schema fields.
// Refer to your definition yaml for the mapping.
type MapModifier[P any] func(r RawMap, plan *P) error

// ComposeModifiers combines multiple MapModifier functions into a single one.
func ComposeModifiers[P any](modifiers ...MapModifier[P]) MapModifier[P] {
	return func(r RawMap, plan *P) error {
		for i, modifier := range modifiers {
			err := modifier(r, plan)
			if err != nil {
				return fmt.Errorf("error applying modifier %d: %w", i, err)
			}
		}
		return nil
	}
}

// FlattenObjectsToArray gets items values from an array of objects by a key.
// For instance, if the array is: [ { "emails": "test@aiven.io" }, { "emails": "test2@aiven.io" } ]
// It will return: [ "test@aiven.io", "test2@aiven.io" ]
func FlattenObjectsToArray[P any](key string, arrayPath ...string) MapModifier[P] {
	return func(r RawMap, _ *P) error {
		data := r.GetData()
		path := PathStr(arrayPath...)

		// foo.#.bar gets all "bar" values from objects in the "foo" array
		array := gjson.GetBytes(data, fmt.Sprintf("%s.#.%s", path, key))
		if !array.Exists() {
			return nil
		}

		if !array.IsArray() {
			return fmt.Errorf("expected array at %q, got %q", path, array.Raw)
		}

		updated, err := sjson.SetRawBytes(data, path, []byte(array.Raw))
		if err != nil {
			return fmt.Errorf("error setting array at %q: %w", path, err)
		}

		return r.SetData(updated)
	}
}

// ExpandArrayToObjects wraps an array of "values" into an object with the specified key.
// it takes an array of "values" and wraps each into an object with the specified key.
// For example: [ "test@aiven.io", "test2@aiven.io" ] -> [ { "email": "test@aiven.io" }, { "email": "test2@aiven.io" } ]
// The field can be optional in Terraform, but required in the API. Set "required" to true to always include it in the API request.
func ExpandArrayToObjects[P any](required bool, key string, arrayPath ...string) MapModifier[P] {
	return func(r RawMap, _ *P) error {
		path := PathStr(arrayPath...)
		arrayResult := gjson.GetBytes(r.GetData(), path)
		if !arrayResult.Exists() && !required {
			return nil
		}

		// When array is "null", ForEach returns [null], so handle null case separately.
		array := []byte("[]")
		if arrayResult.Type == gjson.Null {
			return r.SetRaw(array, arrayPath...)
		}

		if !arrayResult.IsArray() {
			return fmt.Errorf("expected array at %q, got %q", path, arrayResult.Raw)
		}

		// For each array element: create {key: element}
		var err error
		arrayResult.ForEach(func(_, v gjson.Result) bool {
			var value string
			value, err = sjson.SetRaw("", key, v.Raw)
			if err != nil {
				return false
			}

			array, err = sjson.SetRawBytes(array, "-1", []byte(value))
			return err == nil // continue if no error
		})

		if err != nil {
			return fmt.Errorf("error expanding array at %q: %w", path, err)
		}

		return r.SetRaw(array, arrayPath...)
	}
}

// FlattenMapToKeyValue turns a map into an array of objects.
// { "foo": "bar" } -> [ { "key": "foo", "value": "bar" } ]
func FlattenMapToKeyValue[P any](keyField, valueField string, mapPath ...string) MapModifier[P] {
	return func(r RawMap, _ *P) error {
		path := PathStr(mapPath...)
		target := gjson.GetBytes(r.GetData(), path)
		if !target.Exists() {
			return nil
		}

		if !target.IsObject() {
			return fmt.Errorf("expected object at %q, got %q", path, target.Raw)
		}

		// For each key-value pair in the map, create an object {keyField: key, valueField: value}
		var err error
		array := []byte("[]")
		target.ForEach(func(k, v gjson.Result) bool {
			obj, _ := sjson.Set("", keyField, k.String())
			obj, _ = sjson.SetRaw(obj, valueField, v.Raw)
			array, err = sjson.SetRawBytes(array, "-1", []byte(obj))
			return err == nil // continue if no error
		})

		if err != nil {
			return fmt.Errorf("error flattening array at %q: %w", path, err)
		}

		return r.SetRaw(array, mapPath...)
	}
}

// ExpandKeyValueToMap turns array of objects into a map.
// [{"key": "foo", "value": "bar"}] -> { "foo": "bar" }
// The field can be optional in Terraform, but required in the API. Set "required" to true to always include it in the API request.
func ExpandKeyValueToMap[P any](required bool, keyField, valueField string, arrayPath ...string) MapModifier[P] {
	return func(r RawMap, _ *P) error {
		path := PathStr(arrayPath...)
		arrayResult := gjson.GetBytes(r.GetData(), path)
		if !arrayResult.Exists() && !required {
			return nil
		}

		// When array is "null", ForEach returns [null], so handle null case separately.
		array := []byte("{}")
		if arrayResult.Type == gjson.Null {
			return r.SetRaw(array, arrayPath...)
		}

		if !arrayResult.IsArray() {
			return fmt.Errorf("expected array or null at %q, got %q", path, arrayResult.Raw)
		}

		// For each object in the array, extract key and value and set in the map
		var err error
		dict := []byte("{}")
		arrayResult.ForEach(func(_, item gjson.Result) bool {
			key := item.Get(keyField).String()
			value := item.Get(valueField).Raw
			dict, err = sjson.SetRawBytes(dict, key, []byte(value))
			return err == nil // continue if no error
		})

		if err != nil {
			return fmt.Errorf("error expanding key-value to map at %q: %w", path, err)
		}

		return r.SetRaw(dict, arrayPath...)
	}
}
