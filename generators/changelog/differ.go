package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/ettle/strcase"
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	cl "github.com/aiven/terraform-provider-aiven/tools/changelog"
)

func diffItems(was, have *Item) (*Diff, error) {
	if was == nil || have == nil {
		action := AddDiffAction
		if have == nil {
			action = RemoveDiffAction
			have = was
		}

		return &Diff{
			Action: action,
			Item:   have,
		}, nil
	}

	if cmp.Equal(was, have) {
		return nil, nil
	}

	wasMap, err := toMap(was)
	if err != nil {
		return nil, err
	}

	haveMap, err := toMap(have)
	if err != nil {
		return nil, err
	}

	union := make(map[string]any)
	maps.Copy(union, wasMap)
	maps.Copy(union, haveMap)
	keys := slices.Sorted(maps.Keys(union))
	entries := make([]string, 0, len(keys))
	for _, k := range keys {
		wasValue, wasOk := wasMap[k]
		haveValue, haveOk := haveMap[k]
		if !wasOk || !haveOk {
			continue
		}
		if cmp.Equal(wasValue, haveValue) {
			continue
		}

		entry, err := changeEntry(k, was, have, wasValue, haveValue)
		if err != nil {
			return nil, err
		}
		if entry != "" {
			entries = append(entries, entry)
		}
	}

	description := strings.Join(entries, "; ")
	if description == "" {
		return nil, nil
	}

	return &Diff{
		Action:      ChangeDiffAction,
		Description: description,
		Item:        have,
	}, nil
}

func changeEntry(
	key string,
	was, have *Item,
	wasValue, haveValue any,
) (string, error) {
	switch key {
	case "deprecated":
		if have.Deprecated != "" {
			return "deprecated: " + strings.TrimRight(have.Deprecated, ". "), nil
		}
		return "no longer deprecated", nil
	case "beta":
		if haveValue.(bool) {
			return "now beta", nil
		}
		return "no longer beta", nil
	case "limited availability":
		if haveValue.(bool) {
			return "now limited availability", nil
		}
		return "no longer limited availability", nil
	case "enum":
		return cmpList(wasValue.([]string), haveValue.([]string)), nil
	default:
		entry := fmt.Sprintf(
			"%s was `%s`, now `%s`",
			key,
			strValue(wasValue),
			strValue(haveValue),
		)
		return strings.ReplaceAll(entry, "``", "`"), nil
	}
}

func diffItemMaps(was, have ItemMap) ([]cl.GeneratedEntry, error) {
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
			seen[k] = true

			if skipPrefix != "" && strings.HasPrefix(k, skipPrefix) {
				continue
			}

			skipPrefix = ""
			wasVal, wasOk := wasItems[k]
			haveVal, haveOk := haveItems[k]
			if wasOk != haveOk {
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

	values, _ := parseEnum(item.Description)
	m["enum"] = values
	m["beta"] = hasBeta(item.Description)
	m["limitedAvailability"] = hasLimited(item.Description)
	m["type"] = strValueType(item.Type)
	m["elementType"] = strValueType(item.ElementType)

	delete(m, "description")

	for k, v := range m {
		delete(m, k)
		m[strcase.ToCase(k, strcase.LowerCase, ' ')] = v
	}
	return m, err
}

func serializeDiff(list []*Diff) []cl.GeneratedEntry {
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

	out := make([]cl.GeneratedEntry, 0, len(list))
	seen := make(map[string]struct{})
	for _, d := range list {
		e := generatedEntry(d)
		key := entryKey(e)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool {
		return entryKey(out[i]) < entryKey(out[j])
	})
	return out
}

func entryKey(e cl.GeneratedEntry) string {
	return e.Section + ":" + e.Group + ":" + e.Text
}

func cmpList(was, have []string) string {
	const (
		remove int = 1 << iota
		add
	)

	seen := make(map[string]int)
	for _, v := range was {
		seen[v] = remove
	}
	for _, v := range have {
		seen[v] |= add
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
	sort.Strings(added)
	sort.Strings(removed)

	parts := make([]string, 0, 2)
	if len(added) > 0 {
		parts = append(parts, "add "+userconfig.ListOfCodes("and", added...))
	}
	if len(removed) > 0 {
		parts = append(parts, "remove "+userconfig.ListOfCodes("and", removed...))
	}
	return strings.Join(parts, "; ")
}
