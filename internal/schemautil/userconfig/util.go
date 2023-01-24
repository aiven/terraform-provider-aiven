//nolint:unused
package userconfig

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/aiven/aiven-go-client/tools/exp/dist"
	"gopkg.in/yaml.v3"
)

// SchemaType is a custom type that represents a Terraform schema type.
type SchemaType int

const (
	// ServiceTypes is a constant that represents service schema type.
	ServiceTypes SchemaType = iota

	// IntegrationTypes is a constant that represents integration schema type.
	IntegrationTypes

	// IntegrationEndpointTypes is a constant that represents integration endpoint schema type.
	IntegrationEndpointTypes
)

// cachedRepresentationMaps is a map of cached representation maps.
var cachedRepresentationMaps = make(map[SchemaType]map[string]interface{}, 3)

// cachedRepresentationMapsMutex is a mutex for the cached representation maps.
var cachedRepresentationMapsMutex = sync.Mutex{}

// CachedRepresentationMap is a function that returns a cached representation map.
func CachedRepresentationMap(st SchemaType) (map[string]interface{}, error) {
	if _, ok := map[SchemaType]struct{}{
		ServiceTypes:             {},
		IntegrationTypes:         {},
		IntegrationEndpointTypes: {},
	}[st]; !ok {
		return nil, fmt.Errorf("unknown schema type: %d", st)
	}

	switch st {
	case ServiceTypes:
		return representationToMap(st, dist.ServiceTypes)
	case IntegrationTypes:
		return representationToMap(st, dist.IntegrationTypes)
	case IntegrationEndpointTypes:
		return representationToMap(st, dist.IntegrationEndpointTypes)
	default:
		return nil, fmt.Errorf("unknown schema type %d", st)
	}
}

// representationToMap converts a YAML representation of a Terraform schema to a map.
func representationToMap(st SchemaType, r []byte) (map[string]interface{}, error) {
	cachedRepresentationMapsMutex.Lock()
	defer cachedRepresentationMapsMutex.Unlock()

	if v, ok := cachedRepresentationMaps[st]; ok {
		return v, nil
	}

	var m map[string]interface{}
	if err := yaml.Unmarshal(r, &m); err != nil {
		return nil, err
	}

	cachedRepresentationMaps[st] = m
	return m, nil
}

// TerraformTypes is a function that converts schema representation types to Terraform types.
func TerraformTypes(t []string) ([]string, []string, error) {
	var r, ar []string

	for _, v := range t {
		switch v {
		case "null":
			// TODO: We should probably handle this case.
			//  This is a special case where the value can be null.
			//  There should be a default value set for this case.
			continue
		case "boolean":
			r = append(r, "TypeBool")
		case "integer":
			r = append(r, "TypeInt")
		case "number":
			r = append(r, "TypeFloat")
		case "string":
			r = append(r, "TypeString")
		case "array", "object":
			r = append(r, "TypeList")
		default:
			return nil, nil, fmt.Errorf("unknown type: %s", v)
		}

		ar = append(ar, v)
	}

	return r, ar, nil
}

// isTerraformTypePrimitive is a function that checks if a Terraform type is a primitive type.
func isTerraformTypePrimitive(t string) bool {
	switch t {
	case "TypeBool", "TypeInt", "TypeFloat", "TypeString":
		return true
	default:
		return false
	}
}

// mustStringSlice is a function that converts an interface to a slice of strings.
func mustStringSlice(v interface{}) ([]string, error) {
	va, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("not a slice: %#v", v)
	}

	r := make([]string, len(va))

	for k, v := range va {
		va, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("value is not a string: %#v", v)
		}

		r[k] = va
	}

	return r, nil
}

// SlicedString is a function that accepts a string or a slice of strings and returns a slice of strings.
func SlicedString(v interface{}) []string {
	va, ok := v.([]interface{})
	if ok {
		vas, err := mustStringSlice(va)
		if err != nil {
			panic(err)
		}

		return vas
	}

	vsa, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("value is not a string or a slice of strings: %#v", v))
	}

	return []string{vsa}
}

// constDescriptionReplaceables is a slice of strings that are replaced in descriptions.
var constDescriptionReplaceables = map[string]string{
	"DEPRECATED: ":                 "",
	"This setting is deprecated. ": "",
	// TODO: Introduce in a separate PR.
	//"[seconds]":                    "seconds",
}

// descriptionForProperty is a function that returns the description for a property.
func descriptionForProperty(p map[string]interface{}, t string) (id bool, d string) {
	if da, ok := p["description"].(string); ok {
		d = da
	} else {
		d = p["title"].(string)
	}

	if strings.Contains(strings.ToLower(d), "deprecated") {
		id = true
	}

	// Some descriptions have a built-in deprecation notice, so we need to remove it.
	for k, v := range constDescriptionReplaceables {
		d = strings.ReplaceAll(d, k, v)
	}

	b := Desc(d)

	if def, ok := p["default"]; ok && isTerraformTypePrimitive(t) {
		skip := false

		if adef, ok := def.(string); ok {
			if adef == "" {
				skip = true
			}
		}

		if !skip {
			b = b.DefaultValue(def)
		}
	}

	d = b.Build()

	return id, d
}

// EncodeKey is a function that encodes a key for a Terraform schema.
func EncodeKey(k string) string {
	return strings.ReplaceAll(k, ".", "__dot__")
}

// DecodeKey is a function that decodes a key for a Terraform schema.
func DecodeKey(k string) string {
	return strings.ReplaceAll(k, "__dot__", ".")
}

// IsKeyTyped is a function that checks if a key is typed, i.e. has a type suffix in it.
func IsKeyTyped(k string) bool {
	return regexp.MustCompile(`^.*_(boolean|integer|number|string|array|object)$`).MatchString(k)
}
