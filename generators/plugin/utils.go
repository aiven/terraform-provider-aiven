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

// cutDeprecated splits a description with the "DEPRECATED:" prefix, e.g., "DEPRECATED: use index_patterns instead",
// into the title, which describes the field then, e.g., "(DEPRECATED) Maximum index count", and the deprecation message.
func cutDeprecated(description, title string) (string, string, bool) {
	msg, ok := strings.CutPrefix(description, "DEPRECATED:")
	if !ok {
		return description, "", false
	}

	msg = strings.TrimSpace(msg)
	title = strings.TrimSpace(strings.TrimPrefix(title, "(DEPRECATED)"))

	// The docs render the message as is. An empty one would drop the deprecation.
	return or(title, msg), or(userconfig.Desc(msg).Build(), "This field is deprecated."), true
}

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
	return descriptionBuilder(def, entity, item).Build()
}

func descriptionBuilder(def *Definition, entity entityType, item *Item) *userconfig.DescriptionBuilder {
	b := userconfig.Desc(item.Description)
	if enum := itemEnumValues(item); len(enum) > 0 {
		b.PossibleValuesString(schemautil.FlattenToString(enum)...)
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

		// Length hints apply to strings only. Guard by type so a scalar field
		// overridden to an array/set does not inherit a stray "length" hint.
		if item.Type == SchemaTypeString && !hasStringFormat(item) {
			if item.MinLength > 0 {
				b.MinLen(item.MinLength)
			}

			if item.MaxLength > 0 {
				b.MaxLen(item.MaxLength)
			}
		}

		if f, ok := stringFormats[itemFormat(item)]; ok && !f.documented.MatchString(item.Description) {
			b.FormatHint(f.hint)
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

	switch {
	case item.Default != nil:
		b.DefaultValue(item.Default)
	case item.DocDefault != nil:
		b.DefaultValue(item.DocDefault)
	default:
		if v := descExample(item); v != nil {
			b.Example(v)
		}
	}

	// Root deprecation is rendered in doc.go (~> callout); attributes keep the
	// inline **Deprecated**: suffix from DescriptionBuilder.
	if item.DeprecationMessage != "" && !item.IsRoot() {
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

	return b
}

// descExample returns the scalar example of the field: of the definition or the operation, then of the field schema.
// Skips the examples that tell nothing new: of booleans, enums, and fields whose description has the default.
func descExample(item *Item) any {
	switch {
	// todo: drop to document the examples of all fields, regenerating the docs in a separate PR.
	case !item.InUserConfig():
		return nil
	case !item.IsScalar(), item.Type == SchemaTypeBoolean, len(itemEnumValues(item)) > 0,
		strings.Contains(strings.ToLower(item.Description), "defaults to"):
		return nil
	}
	for _, v := range []any{item.Example, item.DocExample} {
		// The spec's float64 rounds an int64 maximum up, e.g., 9223372036854775807 to 9223372036854775808
		if f, ok := v.(float64); ok && item.Maximum > 0 && f >= float64(item.Maximum) {
			v = item.Maximum
		}
		switch v.(type) {
		case string, int, float64:
			if !isEmpty(v) {
				return v
			}
		}
	}
	return nil
}

func itemEnumValues(item *Item) []any {
	if len(item.Enum) > 0 {
		return item.Enum
	}
	if item.Items != nil && len(item.Items.Enum) > 0 {
		return item.Items.Enum
	}
	return item.DocEnum
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
