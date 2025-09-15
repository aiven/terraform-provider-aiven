package main

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

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

func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	return keys
}

var reNewline = regexp.MustCompile(`\s*\n+\s*`)

func fmtDescription(isResource bool, item *Item) string {
	description := strings.TrimSpace(reNewline.ReplaceAllString(item.Description, " "))
	if isResource && !item.IsRoot() && item.Required && item.IsNested() {
		// The documentation generator renders required nested blocks as optional.
		// https://github.com/hashicorp/terraform-plugin-docs/issues/363
		// fixme: remove this once the we render the docs on our own
		description = "Required property. " + description
	}

	b := userconfig.Desc(description)
	if len(item.Enum) > 0 {
		b.PossibleValuesString(schemautil.FlattenToString(item.Enum)...)
	}

	if !isResource {
		b.MarkAsDataSource()
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

func joinCommaAnd(args []string) string {
	switch len(args) {
	case 0:
		return ""
	case 1:
		return args[0]
	}
	return strings.Join(args[:len(args)-1], ", ") + " and " + args[len(args)-1]
}
