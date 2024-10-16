package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
)

func diffItems(resourceType ResourceType, was, have *Item) (*Diff, error) {
	// Added or removed
	if was == nil || have == nil {
		action := ChangeTypeAdd
		if have == nil {
			action = ChangeTypeRemove
			have = was
		}

		return &Diff{
			Type:         action,
			ResourceType: resourceType,
			Description:  removeEnum(have.Description),
			Item:         have,
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
				entry = fmt.Sprintf("deprecate: %s", have.Deprecated)
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
		Type:         ChangeTypeChange,
		ResourceType: resourceType,
		Description:  strings.Join(entries, ", "),
		Item:         have,
	}, nil
}

func diffItemMaps(was, have ItemMap) ([]string, error) {
	result := make([]string, 0)
	kinds := []ResourceType{ResourceKind, DataSourceKind}
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
				result = append(result, change.String())
			}
		}
	}
	return result, nil
}

type DiffType string

const (
	ChangeTypeAdd    DiffType = "Add"
	ChangeTypeRemove DiffType = "Remove"
	ChangeTypeChange DiffType = "Change"
)

type Diff struct {
	Type         DiffType
	ResourceType ResourceType
	Description  string
	Item         *Item
}

func (c *Diff) String() string {
	// resource name + field name
	path := strings.SplitN(c.Item.Path, ".", 2)

	// e.g.: "Add resource `aiven_project`"
	msg := fmt.Sprintf("%s %s `%s`", c.Type, c.ResourceType, path[0])

	// e.g.: "field `project`"
	if len(path) > 1 {
		msg = fmt.Sprintf("%s field `%s`", msg, path[1])
	}

	// Adds beta if needed
	if hasBeta(c.Description) {
		msg = fmt.Sprintf("%s _(beta)_", msg)
	}

	// Adds description
	const maxSize = 120

	msg += ": "
	msg += shorten(maxSize-len(msg), c.Description)
	return msg
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
	m["elemType"] = strValueType(item.ElemType)
	delete(m, "description") // Not needed to compare descriptions
	return m, err
}
