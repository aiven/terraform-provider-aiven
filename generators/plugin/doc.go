package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aiven/go-client-codegen/handler/accountteam"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const (
	docsResourcesDir   = "docs-plugin/resources"
	docsDataSourcesDir = "docs-plugin/data-sources"
	templatesDir       = "templates"
	providerDocName    = "terraform-provider-aiven"
)

type docGroup struct {
	topTitle    string
	nestedTitle string
}

var docGroups = []docGroup{
	{"### Required", "Required:"},
	{"### Optional", "Optional:"},
	{"### Read-Only", "Read-Only:"},
}

type docNestedType struct {
	anchorID  string
	pathTitle string
	path      []string
	item      *Item
	timeouts  bool
}

func docEntityLabel(entity entityType) string {
	if entity.isResource() {
		return "Resource"
	}
	return "Data Source"
}

func docFileName(def *Definition) string {
	return strings.TrimPrefix(def.typeName, typeNamePrefix) + ".md"
}

func docOutputDir(entity entityType) string {
	if entity.isResource() {
		return docsResourcesDir
	}
	return docsDataSourcesDir
}

func hasDocTemplateOverride(def *Definition, entity entityType) bool {
	entityDir := "resources"
	if entity.IsDataSource() {
		entityDir = "data-sources"
	}
	path := filepath.Join(templatesDir, entityDir, docFileName(def)+".tmpl")
	_, err := os.Stat(path)
	return err == nil
}

func importIDPath(def *Definition, root *Item) []string {
	idPath := def.IDAttributeComposed
	if len(idPath) == 0 {
		idPath = []string{root.Name + "_id"}
	}
	return idPath
}

func importID(def *Definition, root *Item) string {
	return strings.ToUpper(filepath.Join(importIDPath(def, root)...))
}

func genDoc(def *Definition, entity entityType, renderRoot *Item, resDat *SchemaMeta) ([]byte, error) {
	if hasDocTemplateOverride(def, entity) {
		return nil, nil
	}

	var b strings.Builder

	description := docRootDescription(def, entity, renderRoot)

	docWriteFrontMatter(&b, def, entity, description)

	b.WriteString("# ")
	b.WriteString(def.typeName)
	b.WriteString(" (")
	b.WriteString(docEntityLabel(entity))
	b.WriteString(")\n\n")
	b.WriteString(strings.TrimSpace(description))
	b.WriteString("\n")

	if example, ok, err := docExample(def, entity, renderRoot, resDat); err != nil {
		return nil, err
	} else if ok {
		b.WriteString("\n## Example Usage\n\n")
		b.WriteString(example)
	}

	schema, err := docSchema(def, entity, renderRoot)
	if err != nil {
		return nil, err
	}
	b.WriteString("\n\n")
	b.WriteString(schema)

	if imp, ok := docImport(def, entity, renderRoot); ok {
		b.WriteString("\n\n")
		b.WriteString(imp)
	}

	return []byte(docFinalize(b.String())), nil
}

func docEntityType(entity entityType) userconfig.EntityType {
	if entity.IsDataSource() {
		return userconfig.DataSource
	}
	return userconfig.Resource
}

// docRootDescription is the resource/data-source page description: schema text
// from fmtDescription plus, for deprecated root entities, the ~> callout that
// tfplugindocs renders from DeprecationMessage (not inlined in MarkdownDescription).
func docRootDescription(def *Definition, entity entityType, renderRoot *Item) string {
	description := strings.TrimSpace(fmtDescription(def, entity, renderRoot))
	if renderRoot.DeprecationMessage == "" {
		return description
	}
	et := docEntityType(entity)
	if description == "" {
		return fmt.Sprintf("~> **This %s is deprecated**\n%s", et, renderRoot.DeprecationMessage)
	}
	return fmt.Sprintf("%s\n\n~> **This %s is deprecated**\n%s", description, et, renderRoot.DeprecationMessage)
}

func docWriteFrontMatter(b *strings.Builder, def *Definition, entity entityType, description string) {
	plain := strings.TrimSpace(plainText(description))

	b.WriteString("---\n")
	b.WriteString("page_title: \"")
	b.WriteString(def.typeName)
	b.WriteString(" ")
	b.WriteString(docEntityLabel(entity))
	b.WriteString(" - ")
	b.WriteString(providerDocName)
	b.WriteString("\"\n")
	b.WriteString("subcategory: \"\"\n")
	b.WriteString("description: |-\n")
	if plain != "" {
		b.WriteString("  ")
		b.WriteString(strings.ReplaceAll(plain, "\n", "\n  "))
		b.WriteByte('\n')
	}
	b.WriteString("---\n\n")
}

func docExample(def *Definition, entity entityType, renderRoot *Item, resDat *SchemaMeta) (string, bool, error) {
	entityString := strings.ReplaceAll(entity.String(), "data", "data-")
	examplePath := filepath.Join(examplesDir, entityString+"s", def.typeName, entityString+".tf")
	content, err := os.ReadFile(examplePath)
	if err != nil && !os.IsNotExist(err) {
		return "", false, fmt.Errorf("read example %s: %w", examplePath, err)
	}
	if err == nil {
		trimmed := strings.TrimSpace(string(content))
		if trimmed != "" {
			return "```terraform\n" + trimmed + "\n```", true, nil
		}
	}

	if resDat.DisableExample {
		return "", false, nil
	}

	exampleBytes, err := exampleRoot(def, entity, renderRoot)
	if err != nil {
		return "", false, err
	}
	generated := strings.TrimSpace(string(exampleBytes))
	if generated == "" {
		return "", false, nil
	}
	return "```terraform\n" + generated + "\n```", true, nil
}

func docImport(def *Definition, entity entityType, root *Item) (string, bool) {
	if !entity.isResource() {
		return "", false
	}
	return fmt.Sprintf("## Import\n\nImport is supported using the following syntax:\n\n```shell\nterraform import %s.example %s\n```",
		def.typeName, importID(def, root)), true
}

func docSchema(def *Definition, entity entityType, root *Item) (string, error) {
	var b strings.Builder
	b.WriteString("## Schema\n\n")
	if err := docWriteBlockChildren(&b, def, entity, root, nil, true); err != nil {
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func docSchemaContainer(item *Item) *Item {
	if item.Items != nil && (item.IsArray() || item.IsMap()) {
		return item.Items
	}
	return item
}

func docPropertyNames(def *Definition, entity entityType, item *Item, root bool) []string {
	props := docSchemaContainer(item).PropertiesByEntity(entity)
	names := make([]string, 0, len(props)+1)
	for k := range props {
		names = append(names, k)
	}
	if root && def != nil {
		names = append(names, "timeouts")
	}
	sort.Strings(names)
	return names
}

func docItemGroupIndex(def *Definition, entity entityType, item *Item, name string) int {
	if name == "timeouts" {
		return 1
	}
	switch {
	case item.IsRequired(def, entity):
		return 0
	case item.IsOptional(def, entity):
		return 1
	case item.IsReadOnly(def, entity):
		return 2
	default:
		panic(fmt.Sprintf("no doc group for %q", name))
	}
}

func docWriteBlockChildren(b *strings.Builder, def *Definition, entity entityType, item *Item, parents []string, root bool) error {
	names := docPropertyNames(def, entity, item, root)
	groups := make([][]string, len(docGroups))

	for _, n := range names {
		if n == "timeouts" {
			groups[1] = append(groups[1], n)
			continue
		}
		child := docSchemaContainer(item).PropertiesByEntity(entity)[n]
		idx := docItemGroupIndex(def, entity, child, n)
		groups[idx] = append(groups[idx], n)
	}

	nestedTypes := make([]docNestedType, 0)

	for i, gf := range docGroups {
		sortedNames := groups[i]
		if len(sortedNames) == 0 {
			continue
		}

		groupTitle := gf.topTitle
		if !root {
			groupTitle = gf.nestedTitle
		}
		b.WriteString(groupTitle)
		b.WriteString("\n\n")

		if docGroupHasWriteOnly(def, entity, item, sortedNames) {
			b.WriteString("> **NOTE**: [Write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments) are supported in Terraform 1.11 and later.\n\n")
		}

		for _, name := range sortedNames {
			path := append(append([]string{}, parents...), name)
			if name == "timeouts" {
				b.WriteString("- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))\n")
				nestedTypes = append(nestedTypes, docNestedType{
					anchorID:  "nestedblock--timeouts",
					pathTitle: "timeouts",
					path:      []string{"timeouts"},
					timeouts:  true,
				})
				continue
			}

			child := docSchemaContainer(item).PropertiesByEntity(entity)[name]
			nt, err := docWritePropertyLine(b, def, entity, path, child)
			if err != nil {
				return fmt.Errorf("render %q: %w", name, err)
			}
			nestedTypes = append(nestedTypes, nt...)
		}

		b.WriteByte('\n')
	}

	return docWriteNestedTypes(b, def, entity, nestedTypes)
}

func docWritePropertyLine(b *strings.Builder, def *Definition, entity entityType, path []string, item *Item) ([]docNestedType, error) {
	name := path[len(path)-1]
	b.WriteString("- `")
	b.WriteString(name)
	b.WriteString("` ")

	typeLabel, err := docTypeLabel(def, entity, item, name)
	if err != nil {
		return nil, err
	}
	b.WriteString(typeLabel)

	desc := strings.TrimSpace(docAttrDescription(def, entity, item))
	if desc != "" {
		b.WriteByte(' ')
		b.WriteString(desc)
	}

	nestedTypes := make([]docNestedType, 0)
	if docHasNestedSchema(item) {
		anchorID := docAnchorID(path, item)
		b.WriteString(" (see [below for nested schema](#")
		b.WriteString(anchorID)
		b.WriteString("))")
		nestedTypes = append(nestedTypes, docNestedType{
			anchorID:  anchorID,
			pathTitle: strings.Join(path, "."),
			path:      path,
			item:      item,
		})
	}

	b.WriteByte('\n')
	return nestedTypes, nil
}

func docHasNestedSchema(item *Item) bool {
	return item.IsNested() || item.IsMapNested()
}

func docAnchorID(path []string, item *Item) string {
	prefix := "nestedblock--"
	if item.IsMapNested() {
		prefix = "nestedatt--"
	}
	return prefix + strings.Join(path, "--")
}

func docTypeLabel(def *Definition, entity entityType, item *Item, name string) (string, error) {
	var b strings.Builder
	b.WriteByte('(')

	switch {
	case item.IsMapNested():
		b.WriteString("Attributes Map")
	case item.IsNested():
		if item.IsSet() {
			b.WriteString("Block Set")
		} else {
			b.WriteString("Block List")
		}
		if entity.isResource() && (item.IsObject() || item.MaxItems > 0) {
			if item.IsObject() && item.MaxItems > 0 {
				return "", fmt.Errorf("object with maxItems > 0")
			}

			// Objects have maxItems=1
			b.WriteString(fmt.Sprintf(", Max: %d", max(1, item.MaxItems)))
		}
	case item.Items != nil && item.IsArray():
		if item.IsSet() {
			b.WriteString("Set of ")
		} else {
			b.WriteString("List of ")
		}
		b.WriteString(docScalarTypeWord(item.Items))
	case item.IsMap():
		b.WriteString("Map of ")
		b.WriteString(docScalarTypeWord(item.Items))
	default:
		b.WriteString(docScalarTypeWord(item))
	}

	if item.Sensitive {
		b.WriteString(", Sensitive")
	}
	if item.WriteOnly {
		b.WriteString(", [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)")
	}
	if item.DeprecationMessage != "" {
		b.WriteString(", Deprecated")
	}

	b.WriteByte(')')
	return b.String(), nil
}

func docScalarTypeWord(item *Item) string {
	switch item.TFType() {
	case "Bool":
		return "Boolean"
	case "Int64", "Float64":
		return "Number"
	default:
		return item.TFType()
	}
}

func docWriteNestedTypes(b *strings.Builder, def *Definition, entity entityType, nestedTypes []docNestedType) error {
	for _, nt := range nestedTypes {
		b.WriteString("<a id=\"")
		b.WriteString(nt.anchorID)
		b.WriteString("\"></a>\n")
		b.WriteString("### Nested Schema for `")
		b.WriteString(nt.pathTitle)
		b.WriteString("`\n\n")

		if nt.timeouts {
			docWriteTimeoutsSection(b, def, entity)
		} else {
			if err := docWriteBlockChildren(b, def, entity, nt.item, nt.path, false); err != nil {
				return err
			}
		}

		b.WriteByte('\n')
	}
	return nil
}

func docWriteTimeoutsSection(b *strings.Builder, def *Definition, entity entityType) {
	b.WriteString("Optional:\n\n")

	type timeoutAttr struct {
		name        string
		deprecated  bool
		description string
	}

	var attrs []timeoutAttr
	if entity.isResource() {
		attrs = []timeoutAttr{
			{
				name:        "create",
				description: "A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\". Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours).",
			},
		}
		if def.LegacyTimeouts {
			attrs = append(attrs, timeoutAttr{
				name:        "default",
				deprecated:  true,
				description: "Timeout for all operations. Deprecated, use operation-specific timeouts instead.",
			})
		}
		attrs = append(attrs,
			timeoutAttr{
				name:        "delete",
				description: "A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\". Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.",
			},
			timeoutAttr{
				name:        "read",
				description: "A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\". Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.",
			},
			timeoutAttr{
				name:        "update",
				description: "A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\". Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours).",
			},
		)
	} else {
		attrs = []timeoutAttr{
			{
				name:        "read",
				description: "A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\". Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours).",
			},
		}
	}

	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].name < attrs[j].name
	})

	for _, attr := range attrs {
		b.WriteString("- `")
		b.WriteString(attr.name)
		b.WriteString("` (String")
		if attr.deprecated {
			b.WriteString(", Deprecated)")
		} else {
			b.WriteByte(')')
		}
		if attr.description != "" {
			b.WriteByte(' ')
			b.WriteString(attr.description)
		}
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
}

func docAttrDescription(def *Definition, entity entityType, item *Item) string {
	desc := fmtDescription(def, entity, item)
	// Mirror runtime schema patches applied in hand-written resource code.
	if def.typeName == "aiven_organization_permission" && entity.isResource() && item.Path() == "permissions/permissions" {
		desc = userconfig.Desc(desc).PossibleValuesString(accountteam.TeamTypeChoices()...).Build()
	}
	return desc
}

func docGroupHasWriteOnly(def *Definition, entity entityType, item *Item, names []string) bool {
	props := docSchemaContainer(item).PropertiesByEntity(entity)
	for _, name := range names {
		if name == "timeouts" {
			continue
		}
		child, ok := props[name]
		if ok && child.WriteOnly {
			return true
		}
	}
	return false
}

func docFinalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n") + "\n"
}
