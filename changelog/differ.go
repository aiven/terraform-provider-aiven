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

func diffItems(was, have *Item) (*Diff, error) {
	// Added or removed
	if was == nil || have == nil {
		action := AddDiffAction
		if have == nil {
			action = RemoveDiffAction
			have = was
		}

		return &Diff{
			Action:      action,
			Description: removeEnum(have.Description), // Some fields have enums in the spec description
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
		case "enum":
			entry = cmpList(wasValue.([]string), haveValue.([]string))
		default:
			// The rest of the fields will have diff-like entry
			entry = fmt.Sprintf("%s ~~`%s`~~ → `%s`", k, strValue(wasValue), strValue(haveValue))

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
		Description: strings.Join(entries, ", "),
		Item:        have,
	}, nil
}

func diffItemMaps(was, have ItemMap) ([]string, error) {
	result := make([]*Diff, 0)
	kinds := []RootKind{ResourceRootKind, DataSourceRootKind}
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

			change, err := diffItems(wasVal, haveVal)
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

		if a.Item.Root != b.Item.Root {
			return a.Item.Root < b.Item.Root
		}

		if a.Action != b.Action {
			return a.Action < b.Action
		}

		if a.Item.Path != b.Item.Path {
			return a.Item.Path < b.Item.Path
		}

		if a.Item.Kind != b.Item.Kind {
			return a.Item.Kind > b.Item.Kind
		}

		return false
	})

	// Removes duplicates
	unique := make(map[string]*Diff)
	for i := 0; i < len(list); i++ {
		d := list[i]
		k := fmt.Sprintf("%s:%s:%s", d.Action, d.Item.Path, d.Description)
		other, ok := unique[k]
		if !ok {
			unique[k] = d
			continue
		}
		other.AlsoAppliesTo = d.Item
		list = append(list[:i], list[i+1:]...)
		i--
	}

	strs := make([]string, len(list))
	for i, r := range list {
		strs[i] = r.String()
	}
	return strs
}

func cmpList[T any](was, have []T) string {
	const (
		remove int = 1 << iota
		add
	)

	seen := make(map[string]int)
	for _, v := range was {
		seen[fmt.Sprintf("`%v`", v)] = remove
	}

	for _, v := range have {
		seen[fmt.Sprintf("`%v`", v)] |= add
	}

	var added, removed []string
	for k, v := range seen {
		switch v {
		case add:
			added = append(added, k)
		case remove:
			removed = append(removed, k)
		}
	}

	result := make([]string, 0)
	if s := joinSorted(added); s != "" {
		result = append(result, "add "+s)
	}

	if s := joinSorted(removed); s != "" {
		result = append(result, "remove "+s)
	}

	return joinSorted(result)
}

func joinSorted(args []string) string {
	sort.Strings(args)
	return strings.Join(args, ", ")
}
