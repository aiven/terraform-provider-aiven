package main

import (
	"fmt"
	"strings"

	cl "github.com/aiven/terraform-provider-aiven/tools/changelog"
)

func generatedEntry(diff *Diff) cl.GeneratedEntry {
	verb, section := verbSection(diff.Action)
	item := diff.Item

	if group, rel, ok := userConfigParts(item.Path); ok {
		return cl.GeneratedEntry{
			Section: cl.SectionUserConfigurations,
			Group:   group,
			Text:    cl.NormalizeSpace(verb + " " + formatRelPath(rel, item, diff)),
		}
	}

	return cl.GeneratedEntry{
		Section: section,
		Text:    cl.NormalizeSpace(formatFlatEntry(item, diff)),
	}
}

func verbSection(action DiffAction) (verb, section string) {
	switch action {
	case AddDiffAction:
		return "Added", cl.SectionNew
	case RemoveDiffAction:
		return "Removed", cl.SectionImproved
	default:
		return "Changed", cl.SectionImproved
	}
}

func userConfigParts(path string) (group, rel string, ok bool) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 || !strings.HasSuffix(parts[1], "_user_config") {
		return "", "", false
	}
	group = fmt.Sprintf("`%s.%s`", parts[0], parts[1])
	if len(parts) == 2 {
		rel = parts[1]
	} else {
		rel = strings.Join(parts[2:], ".")
	}
	return group, rel, true
}

func formatFlatEntry(item *Item, diff *Diff) string {
	parts := strings.SplitN(item.Path, ".", 2)
	root := parts[0]

	var head string
	if len(parts) == 1 {
		head = fmt.Sprintf("`%s`%s%s", root, kindLabel(item, diff), markers(item, diff, false))
	} else {
		head = fmt.Sprintf(
			"`%s`%s field `%s`%s",
			root,
			kindLabel(item, diff),
			parts[1],
			markers(item, diff, false),
		)
	}

	detail := formatDetail(diff, false)
	if detail == "" {
		return head
	}
	return head + ": " + detail
}

func kindLabel(item *Item, diff *Diff) string {
	if diff.AlsoAppliesTo != nil {
		return " resource and data source"
	}
	switch item.Kind {
	case ResourceRootKind:
		return " resource"
	case DataSourceRootKind:
		return " data source"
	default:
		return ""
	}
}

func markers(item *Item, diff *Diff, userConfigChild bool) string {
	out := parentheticalTags(item, diff, userConfigChild)
	if diff.Action == ChangeDiffAction {
		return out
	}
	if hasBeta(item.Description) {
		out += " _(beta)_"
	}
	if hasLimited(item.Description) {
		out += " _(limited availability)_"
	}
	return out
}

func parentheticalTags(item *Item, diff *Diff, userConfigChild bool) string {
	var parts []string
	if isEnum(item.Description) {
		parts = append(parts, "enum")
	}
	if userConfigChild && diff.AlsoAppliesTo == nil && diff.Action == ChangeDiffAction {
		switch item.Kind {
		case ResourceRootKind:
			parts = append(parts, "resource only")
		case DataSourceRootKind:
			parts = append(parts, "data source only")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

func formatRelPath(rel string, item *Item, diff *Diff) string {
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(rel)
	b.WriteString("`")
	b.WriteString(markers(item, diff, true))
	detail := formatDetail(diff, true)
	if detail != "" {
		b.WriteString(": ")
		b.WriteString(detail)
	}
	return b.String()
}

func formatDetail(diff *Diff, userConfigChild bool) string {
	switch diff.Action {
	case RemoveDiffAction:
		if userConfigChild {
			return ""
		}
		return "removed"
	case AddDiffAction:
		return addDetail(diff.Item.Description)
	default:
		return diff.Description
	}
}
