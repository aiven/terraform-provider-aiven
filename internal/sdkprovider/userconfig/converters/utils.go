package converters

import (
	"fmt"
	"sort"
	"strings"
)

// sortByKey sorts the given array of objects by values in the original array by the given key.
// For instance, when ip_filter_object list is sent, it is sorted on the backend.
// That makes a diff, because user defined order is violated.
func sortByKey(sortBy string, originalSrc, dtoSrc any) any {
	original := asMapList(originalSrc)
	dto := asMapList(dtoSrc)
	if len(original) != len(dto) {
		return dtoSrc
	}

	sortMap := make(map[string]int)
	for i, v := range original {
		sortMap[v[sortBy].(string)] = i
	}

	sort.Slice(dto, func(i, j int) bool {
		ii := dto[i][sortBy].(string)
		jj := dto[j][sortBy].(string)
		return sortMap[ii] > sortMap[jj]
	})

	// Need to cast to "any",
	// otherwise it might blow up in flattenObj function
	// with type mismatch (map[string]any vs any)
	result := make([]any, 0, len(dto))
	for _, v := range dto {
		result = append(result, v)
	}
	return result
}

// drillKey gets deep down key value
func drillKey(dto map[string]any, path string) (any, bool) {
	if dto == nil {
		return nil, false
	}

	keys := strings.Split(path, ".0.")
	keysLen := len(keys) - 1
	for i := 0; ; i++ {
		v, ok := dto[keys[i]]
		if !ok {
			return nil, false
		}

		isLast := i == keysLen
		if isLast {
			return v, true
		}

		next, ok := v.(map[string]any)
		if ok {
			dto = next
			continue
		}

		// Gets the first element of an array
		list, ok := v.([]any)
		if !ok || len(list) == 0 {
			return nil, false
		}

		next, ok = list[0].(map[string]any)
		if !ok {
			return nil, false
		}
		dto = next
	}
}

// asList converts "any" to specific typed list
func asList[T any](v any) []T {
	list := v.([]any)
	result := make([]T, 0, len(list))
	for _, item := range list {
		result = append(result, item.(T))
	}
	return result
}

// asMapList converts "any" to a list of objects
func asMapList(v any) []map[string]any {
	return asList[map[string]any](v)
}

// castType returns an error on invalid type
func castType[T any](v any) (T, error) {
	t, ok := v.(T)
	if !ok {
		var empty T
		return empty, fmt.Errorf("invalid type. Expected %T, got %T", empty, v)
	}
	return t, nil
}
