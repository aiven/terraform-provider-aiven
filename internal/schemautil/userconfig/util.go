//nolint:unused
package userconfig

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aiven/aiven-go-client/tools/exp/dist"
	"gopkg.in/yaml.v3"
)

// SchemaType is a custom type that represents a Terraform schema type.
type SchemaType int

const (
	ServiceTypes SchemaType = iota
	IntegrationTypes
	IntegrationEndpointTypes
)

// cachedRepresentationMaps is a map of cached representation maps.
var cachedRepresentationMaps = make(map[SchemaType]map[string]interface{}, 3)

// CachedRepresentationMap is a function that returns a cached representation map.
func CachedRepresentationMap(st SchemaType) map[string]interface{} {
	if _, ok := map[SchemaType]struct{}{
		ServiceTypes:             {},
		IntegrationTypes:         {},
		IntegrationEndpointTypes: {},
	}[st]; !ok {
		panic(fmt.Sprintf("unknown schema type: %d", st))
	}

	if _, ok := cachedRepresentationMaps[st]; !ok {
		switch st {
		case ServiceTypes:
			representationToMap(st, dist.ServiceTypes)
		case IntegrationTypes:
			representationToMap(st, dist.IntegrationTypes)
		case IntegrationEndpointTypes:
			representationToMap(st, dist.IntegrationEndpointTypes)
		}
	}

	return cachedRepresentationMaps[st]
}

// representationToMap converts a YAML representation of a Terraform schema to a map.
func representationToMap(st SchemaType, r []byte) map[string]interface{} {
	var m map[string]interface{}

	if err := yaml.Unmarshal(r, &m); err != nil {
		panic(err)
	}

	cachedRepresentationMaps[st] = m

	return m
}

// TerraformTypes is a function that converts schema representation types to Terraform types.
func TerraformTypes(t []string) ([]string, []string) {
	var r, ar []string

	for _, v := range t {
		switch v {
		case "null":
			// TODO: We should probably handle this case.
			//  This is a special case where the value can be null.
			//  There should be a default value set for this case.
			continue
		case "boolean":
			// TODO: Make those types be actual types instead of strings.
			// r = append(r, "TypeBool")
			r = append(r, "TypeString")
		case "integer":
			// r = append(r, "TypeInt")
			r = append(r, "TypeString")
		case "number":
			// r = append(r, "TypeFloat")
			r = append(r, "TypeString")
		case "string":
			r = append(r, "TypeString")
		case "array", "object":
			r = append(r, "TypeList")
		default:
			panic(fmt.Sprintf("unknown type: %s", v))
		}

		ar = append(ar, v)
	}

	return r, ar
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
func mustStringSlice(v interface{}) []string {
	va, ok := v.([]interface{})
	if !ok {
		panic(fmt.Sprintf("not a slice: %#v", v))
	}

	r := make([]string, len(va))

	for k, v := range va {
		va, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("value is not a string: %#v", v))
		}

		r[k] = va
	}

	return r
}

// SlicedString is a function that accepts a string or a slice of strings and returns a slice of strings.
func SlicedString(v interface{}) []string {
	va, ok := v.([]interface{})
	if ok {
		return mustStringSlice(va)
	}

	vsa, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("value is not a string or a slice of strings: %#v", v))
	}

	return []string{vsa}
}

// descriptionForProperty is a function that returns the description for a property.
func descriptionForProperty(p map[string]interface{}) (string, string) {
	k := "Description"

	if d, ok := p["description"].(string); ok {
		if strings.Contains(strings.ToLower(d), "deprecated") {
			k = "Deprecated"
		}

		return k, d
	}

	return k, p["title"].(string)
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
