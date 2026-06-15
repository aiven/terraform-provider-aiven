package main

import (
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"

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

// isValidRegex reports whether pattern is a non-empty regular expression
// compilable by Go's RE2-based `regexp` package.
// OpenAPI specs may use extended Perl features (lookarounds, backreferences,
// possessive quantifiers, etc.) that RE2 doesn't support; those return false.
func isValidRegex(pattern string) bool {
	if pattern == "" {
		return false
	}
	_, err := regexp.Compile(pattern)
	return err == nil
}

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
		if item.IsRootProperty() {
			switch {
			case slices.Contains(def.Datasource.ExactlyOneOf, item.Name):
				b.ExactlyOneOf(def.Datasource.ExactlyOneOf...)
			case !item.FromSchemaOverride && def.DatasourceLookupHas(item.Name):
				b.Lookup(def.DatasourceLookupID(), def.DatasourceLookupComposedOf()...)
			}
		}
	} else if !item.IsReadOnly(def, entity) {
		if item.ForceNew {
			b.ForceNew()
		}

		if item.MinLength > 0 {
			b.MinLen(item.MinLength)
		}

		if item.MaxLength > 0 {
			b.MaxLen(item.MaxLength)
		}

		if item.Minimum > 0 {
			b.Minimum(item.Minimum)
		}

		if item.Maximum > 0 {
			b.Maximum(item.Maximum)
		}

		if isValidRegex(item.Pattern) {
			b.Pattern(item.Pattern)
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
	// ExactlyOneOf already implies "not together", so suppress overlapping
	// ConflictsWith. Mirrored in genValidators.
	exactly := item.ExactlyOneOf
	if !isResource && item.IsRootProperty() && slices.Contains(def.Datasource.ExactlyOneOf, item.Name) {
		exactly = def.Datasource.ExactlyOneOf
	}
	if conflicts := lo.Without(item.ConflictsWith, exactly...); len(conflicts) > 0 {
		b.ConflictsWith(conflicts...)
	}
	if item.ExactlyOneOf != nil {
		b.ExactlyOneOf(item.ExactlyOneOf...)
	}
	if item.AtLeastOneOf != nil {
		b.AtLeastOneOf(item.AtLeastOneOf...)
	}

	if item.IsRoot() {
		b.IsRoot()

		if lo.FromPtr(def.Beta) {
			b.Beta()
		}
		if lo.FromPtr(def.LimitedAvailability) {
			b.LimitedAvailability()
		}
		if entity.isResource() && def.Resource.RemoveMissing {
			b.RemoveMissing()
		}
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

// isEmpty checks if the value is empty — has length zero.
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}

	return false
}
