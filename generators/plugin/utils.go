package main

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// distinct doesn't preserve the order of the input arguments.
func distinct[T any](args ...T) []T {
	seen := make(map[string]T, len(args))
	for _, v := range args {
		seen[fmt.Sprint(v)] = v
	}

	keys := slices.Collect(maps.Keys(seen))
	slices.Sort(keys)
	list := make([]T, 0, len(seen))
	for _, k := range keys {
		list = append(list, seen[k])
	}
	return list
}

func mergeSlices[T any](args ...[]T) []T {
	merged := make([]T, 0)
	for _, a := range args {
		merged = append(merged, a...)
	}

	result := distinct(merged...)
	if len(result) == 0 {
		return nil
	}
	return result
}

func or[T comparable](a, b T) T {
	var zero T
	if a != zero {
		return a
	}
	return b
}

func orLonger[T ~string](a, b T) T {
	if len(a) > len(b) {
		return a
	}
	return b
}

func ptrOrDefault[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

// sortedKeys sorts the keys, "id" first, then alphabetically
func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := slices.Collect(maps.Keys(m))
	slices.SortFunc(keys, func(i, j K) int {
		if i == "id" {
			return -1
		} else if j == "id" {
			return 1
		}
		return strings.Compare(string(i), string(j))
	})
	return keys
}

var reNewline = regexp.MustCompile(`\s*\n+\s*`)

func fmtDescription(def *Definition, entity entityType, item *Item) string {
	description := strings.TrimSpace(reNewline.ReplaceAllString(item.Description, " "))
	if entity.isResource() && !item.IsRoot() && item.Required && item.IsNested() {
		// The documentation generator renders required nested blocks as optional.
		// https://github.com/hashicorp/terraform-plugin-docs/issues/363
		// fixme: remove this once the we render the docs on our own
		description = "Required property. " + description
	}

	b := userconfig.Desc(description)
	if len(item.Enum) > 0 {
		b.PossibleValuesString(schemautil.FlattenToString(item.Enum)...)
	}

	isResource := entity.isResource()
	if !isResource {
		b.MarkAsDataSource()
		if item.IsRootProperty() && slices.Contains(def.Datasource.ExactlyOneOf, item.Name) {
			b.ExactlyOneOf(def.Datasource.ExactlyOneOf...)
		}
	} else if !item.IsReadOnly(isResource) {
		if item.ForceNew {
			b.ForceNew()
		}

		if item.MaxLength > 0 {
			b.MaxLen(item.MaxLength)
		}
	}

	if isResource && item.ForceNew {
		b.ForceNew()
	}

	if item.Default != nil {
		b.DefaultValue(item.Default)
	}

	if item.DeprecationMessage != "" {
		b.Deprecated(item.DeprecationMessage)
	}

	// Validators
	if item.AlsoRequires != nil {
		b.RequiredWith(item.AlsoRequires...)
	}
	if item.ConflictsWith != nil {
		b.ConflictsWith(item.ConflictsWith...)
	}
	if item.ExactlyOneOf != nil {
		b.ExactlyOneOf(item.ExactlyOneOf...)
	}
	if item.AtLeastOneOf != nil {
		b.AtLeastOneOf(item.AtLeastOneOf...)
	}

	return b.Build()
}

func firstUpper[T ~string](s T) string {
	v := string(s)
	if v == "" {
		return v
	}
	return strings.ToUpper(v[:1]) + v[1:]
}

func dictFromMap(m map[string]jen.Code, litKeys bool) jen.Dict {
	dict := make(jen.Dict)
	for k, v := range m {
		if litKeys {
			dict[jen.Lit(k)] = v
		} else {
			dict[jen.Id(k)] = v
		}
	}
	return dict
}

func multilineCall() jen.Options {
	return jen.Options{
		Close:     ")",
		Multi:     true,
		Open:      "(",
		Separator: ",",
	}
}

func multilineValues() jen.Options {
	return jen.Options{
		Close:     "}",
		Multi:     true,
		Open:      "{",
		Separator: ",",
	}
}
