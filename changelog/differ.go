package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ettle/strcase"
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
)

func diffItems(resourceType RootType, was, have *Item) (*Diff, error) {
	// Added or removed
	if was == nil || have == nil {
		action := AddDiffAction
		if have == nil {
			action = RemoveDiffAction
			have = was
		}

		return &Diff{
			Action:      action,
			RootType:    resourceType,
			Description: removeEnum(have.Description),
			Item:        have,
		}, nil
	}

	// Equal items
	if cmp.Equal(was, have) {
		return nil, nil
	}

	// Compare all the fields
	wasMap, err := toMap(was)
	if err != nil {
		return nil, err
	}

	haveMap, err := toMap(have)
	if err != nil {
		return nil, err
	}

	entries := make([]string, 0)
	for k, wasValue := range wasMap {
		haveValue := haveMap[k]
		if cmp.Equal(wasValue, haveValue) {
			continue
		}

		var entry string
		switch k {
		case "deprecated":
			entry = "remove deprecation"
			if have.Deprecated != "" {
				entry = fmt.Sprintf("deprecate: %s", strings.TrimRight(have.Deprecated, ". "))
			}
		case "beta":
			entry = "marked as beta"
			if !haveValue.(bool) {
				entry = "no longer beta"
			}
		default:
			// The rest of the fields will have diff-like entry
			entry = fmt.Sprintf("%s ~~`%s`~~ -> `%s`", k, strValue(wasValue), strValue(haveValue))

			// Fixes formatting issues
			entry = strings.ReplaceAll(entry, "``", "`")
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, nil
	}

	return &Diff{
		Action:      ChangeDiffAction,
		RootType:    resourceType,
		Description: strings.Join(entries, ", "),
		Item:        have,
	}, nil
}

func diffItemMaps(was, have ItemMap) ([]string, error) {
	result := make([]*Diff, 0)
	kinds := []RootType{ResourceRootType, DataSourceRootType}
	for _, kind := range kinds {
		wasItems := was[kind]
		haveItems := have[kind]
		keys := append(lo.Keys(wasItems), lo.Keys(haveItems)...)
		slices.Sort(keys)

		var skipPrefix string
		seen := make(map[string]bool)
		for _, k := range keys {
			if seen[k] {
				continue
			}

			// Skips duplicate keys
			seen[k] = true

			// When a resource added or removed, it skips all its fields until the next resource
			// Otherwise, all its fields will appear as changes
			if skipPrefix != "" && strings.HasPrefix(k, skipPrefix) {
				continue
			}

			skipPrefix = ""
			wasVal, wasOk := wasItems[k]
			haveVal, haveOk := haveItems[k]
			if wasOk != haveOk {
				// Resource added or removed, must skip all its fields
				skipPrefix = k
			}

			change, err := diffItems(kind, wasVal, haveVal)
			if err != nil {
				return nil, fmt.Errorf("failed to compare %s %s: %w", kind, k, err)
			}

			if change != nil {
				result = append(result, change)
			}
		}
	}

	return serializeDiff(result), nil
}

func toMap(item *Item) (map[string]any, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	m := make(map[string]any)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	m["enum"] = findEnums(item.Description)
	m["beta"] = hasBeta(item.Description)
	m["type"] = strValueType(item.Type)
	m["elementType"] = strValueType(item.ElementType)

	// Not needed to compare descriptions
	delete(m, "description")

	// Turns "maxItems" into "max items" for human readability
	for k, v := range m {
		delete(m, k)
		m[strcase.ToCase(k, strcase.LowerCase, ' ')] = v
	}
	return m, err
}

func serializeDiff(list []*Diff) []string {
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.Action != b.Action {
			return a.Action < b.Action
		}

		// Resource comes first, then datasource
		if a.RootType != b.RootType {
			return a.RootType > b.RootType
		}

		return a.Item.Path < b.Item.Path
	})

	strs := make([]string, len(list))
	for i, r := range list {
		strs[i] = r.String()
	}
	return strs
}