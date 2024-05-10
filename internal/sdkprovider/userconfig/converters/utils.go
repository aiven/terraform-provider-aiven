package converters

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// castType returns an error on invalid type
func castType[T any](v any) (T, error) {
	t, ok := v.(T)
	if !ok {
		var empty T
		return empty, fmt.Errorf("invalid type. Expected %T, got %T", empty, v)
	}
	return t, nil
}

// renameAliasesToDto renames aliases to DTO object
// Must sort keys to rename from bottom to top.
// Otherwise, might not find the deepest key if parent key is renamed
func renameAliasesToDto(kind userConfigType, name string, dto map[string]any) {
	m := getFieldMapping(kind, name)
	for _, from := range sortKeys(m) {
		renameAlias(dto, from, m[from])
	}
}

// resourceData to test schema.ResourceData with unit tests
type resourceData interface {
	GetOk(string) (any, bool)
}

// renameAliasesToTfo renames aliases to TF object
// Must sort keys to rename from bottom to top.
// Otherwise, might not find the deepest key if parent key is renamed
func renameAliasesToTfo(kind userConfigType, name string, dto map[string]any, d resourceData) error {
	prefix := userConfigKey(kind, name) + ".0."
	m := getFieldMapping(kind, name)

	for _, to := range sortKeys(m) {
		from := m[to]

		if strings.HasSuffix(to, "_string") || strings.HasSuffix(to, "_object") {
			// If resource doesn't have this field, then ignores (uses original)
			path := strings.ReplaceAll(to, "/", ".0.")
			_, ok := d.GetOk(prefix + path)
			if !ok {
				continue
			}
		}

		renameAlias(dto, from, to)
	}

	// Converts ip_filter list into an expected by the config type
	return convertIPFilter(dto)
}

// ipFilterMistyped reverse types: string to map, map to string
// Unmarshalled with no errors when ip_filter has type missmatch
type ipFilterMistyped struct {
	IPFilter       []map[string]string `json:"ip_filter"`
	IPFilterString []map[string]string `json:"ip_filter_string"`
	IPFilterObject []string            `json:"ip_filter_object"`
}

// convertIPFilter converts a list of ip_filter objects into a list of strings and vice versa
func convertIPFilter(dto map[string]any) error {
	b, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	var r ipFilterMistyped
	err = json.Unmarshal(b, &r)
	if err != nil {
		// nolint: nilerr
		// Marshaling went wrong, nothing to fix
		return nil
	}

	// Converting went smooth.
	// Which means either there is no ip_filter at all, or it has an invalid type

	// Converts strings into objects
	if len(r.IPFilterObject) > 0 {
		mapList := make([]map[string]string, 0)
		for _, v := range r.IPFilterObject {
			mapList = append(mapList, map[string]string{"network": v})
		}

		dto["ip_filter_object"] = mapList
		return nil
	}

	// Converts objects into strings
	strList := make([]string, 0)
	for _, v := range append(r.IPFilter, r.IPFilterString...) {
		strList = append(strList, v["network"])
	}

	if len(strList) == 0 {
		// Nothing to do here
		return nil
	}

	// Chooses which key to set values to
	strKey := "ip_filter"
	if len(r.IPFilterString) > 0 {
		strKey = "ip_filter_string"
	}

	dto[strKey] = strList
	return nil
}

// renameAlias renames ip_filter_string to ip_filter
func renameAlias(dto map[string]any, from, to string) {
	keys := strings.Split(from, "/")
	orig := strings.Split(to, "/")[len(keys)-1]
	for i, k := range keys {
		v, ok := dto[k]
		if !ok {
			// the key does not exist
			return
		}

		// If reached the key
		isLast := len(keys)-1 == i
		if isLast {
			delete(dto, k)
			if !isZero(dto[orig]) {
				// When there multiple choices, like ip_filter_string and ip_filter_object,
				// keeps the one that has a non-empty value.
				// This might happen when a user migrates from ip_filter to ip_filter_string
				v = dto[orig]
			}
			dto[orig] = v
			return
		}

		if a, ok := v.([]any); ok {
			for _, j := range a {
				renameAlias(j.(map[string]any), strings.Join(keys[i+1:], "/"), orig)
			}
			return
		}
		dto = v.(map[string]any)
	}
}

// isZero returns true for zero-values
func isZero(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Map, reflect.Slice:
		return value.Len() == 0
	}
	return value.IsZero()
}

func sortKeys[T any](m map[string]T) []string {
	keys := maps.Keys(m)
	slices.SortFunc(keys, func(a, b string) int {
		return len(b) - len(a)
	})
	return keys
}
