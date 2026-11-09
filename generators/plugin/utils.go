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

// reParagraph splits a description on blank lines (a newline followed by
// optional whitespace and another newline), i.e. author-intended paragraph
// breaks.
var reParagraph = regexp.MustCompile(`\n\s*\n`)

// descModifier rewrites part of a description so it renders correctly in
// Markdown.
type descModifier struct {
	re   *regexp.Regexp
	repl string
}

// symbolRun matches a run of symbol characters: no letters (\pL), digits (\pN),
// whitespace or quotes. It is spelled out rather than using \W so that "_",
// which \w counts as a word character, is treated as a symbol too. Excluding
// whitespace is what keeps the gap between two quoted values from being
// mistaken for a quoted symbol: a quoted token such as ".*" never contains
// whitespace, while every such gap does (e.g. the ": " in {"a": "b"}).
const symbolRun = `([^\pL\pN\s'"]+)`

// descModifiers are applied in order by applyDescModifiers:
//   - a whitespace-delimited "*" (the wildcard meaning "all hosts") is wrapped
//     in backticks, keeping the surrounding whitespace.
//   - a quoted symbol run (e.g. '.', "/", ".*") has its quotes replaced by
//     backticks, for single quotes, double quotes and the &quot; entity that
//     the OpenAPI spec ships.
var descModifiers = []descModifier{
	{regexp.MustCompile(`(\s+)\*(\s+)`), "${1}`*`${2}"},
	{regexp.MustCompile(`'` + symbolRun + `'`), "`${1}`"},
	{regexp.MustCompile(`"` + symbolRun + `"`), "`${1}`"},
	{regexp.MustCompile(`&quot;` + symbolRun + `&quot;`), "`${1}`"},
}

// applyDescModifiers applies descModifiers in order.
func applyDescModifiers(s string) string {
	for _, m := range descModifiers {
		s = m.re.ReplaceAllString(s, m.repl)
	}
	return s
}

// reVerbatim matches what backtickIdentifiers keeps: code spans, URLs, Markdown link targets
// and "${...}" placeholders, which it wraps whole.
var reVerbatim = regexp.MustCompile("`[^`]*`" + `|https?://\S+|\]\([^)]*\)|\$\{[^}\s]*\}`)

// reIdentifier matches a lowercase or uppercase word or dotted path,
// e.g., "pg_stat_monitor.pgsm_max_buckets" or "PERFORMANCE_SCHEMA.EVENT_WAITS".
var reIdentifier = regexp.MustCompile(`[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*|[A-Z][A-Z0-9_]*(?:\.[A-Z][A-Z0-9_]*)*`)

// reSettingValue matches the value after an identifier, e.g., "=worker" of "io_method=worker".
var reSettingValue = regexp.MustCompile(`^=[^\s,;)'"&]*`)

// backtickIdentifiers wraps the lowercase and uppercase words with "_" in backticks, e.g., sas_token
// and AIVEN_TOKEN, replacing the quotes around them.
func backtickIdentifiers(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range reVerbatim.FindAllStringIndex(s, -1) {
		b.WriteString(backtickProse(s[last:m[0]]))
		if v := s[m[0]:m[1]]; strings.HasPrefix(v, "${") {
			b.WriteString("`" + v + "`")
		} else {
			b.WriteString(v)
		}
		last = m[1]
	}
	b.WriteString(backtickProse(s[last:]))
	return b.String()
}

func backtickProse(s string) string {
	// Part of a longer token, e.g., "Max_size", "a-b_c", "path/to_file", "my_user@host" or "Long.MAX_VALUE"
	glued := func(c byte, symbols string) bool {
		return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte(symbols, c) >= 0
	}

	var b strings.Builder
	last := 0
	for _, m := range reIdentifier.FindAllStringIndex(s, -1) {
		start, end := m[0], m[1]
		switch {
		case start < last, // In the value of a setting
			!strings.Contains(s[start:end], "_"),
			start > 0 && glued(s[start-1], `_-/@.$\`),
			end < len(s) && glued(s[end], `-/@\`):
			continue
		}

		// A setting, e.g., "io_method=worker"
		if v := reSettingValue.FindString(s[end:]); v != "" {
			end += len(strings.TrimRight(v, "."))
		}

		before, after := start, end
		for _, q := range []string{"'", `"`, "&quot;"} {
			if strings.HasSuffix(s[:start], q) && strings.HasPrefix(s[end:], q) {
				before, after = start-len(q), end+len(q)
				break
			}
		}
		// A JSON key or string value, e.g., {"key_name": "value_name"}
		quoted := after > end
		jsonKey := quoted && strings.HasPrefix(s[after:], ":")
		jsonValue := quoted && s[before] != '\'' && strings.HasSuffix(strings.TrimRight(s[:before], " "), ":")
		if jsonKey || jsonValue {
			continue
		}

		b.WriteString(s[last:before])
		b.WriteString("`" + s[start:end] + "`")
		last = after
	}
	b.WriteString(s[last:])
	return b.String()
}

// reJSONObject matches a JSON object without nested ones, e.g., {"key_name":"value"}.
var reJSONObject = regexp.MustCompile(`\{\s*(?:"|&quot;)[^{}]*\}`)

// normalizeDescription unwraps hard-wrapped lines into single spaces and runs
// descModifiers and backtickIdentifiers over the result, except in JSON objects.
// When preserveParagraphs is true (root entity descriptions), blank lines in the
// source are kept as Markdown paragraph breaks; otherwise every newline collapses
// to a space so attribute descriptions stay on one line (docs render them as
// list/table entries).
func normalizeDescription(s string, preserveParagraphs bool) string {
	collapse := func(p string) string {
		p = strings.TrimSpace(reNewline.ReplaceAllString(p, " "))
		var b strings.Builder
		last := 0
		for _, m := range reJSONObject.FindAllStringIndex(p, -1) {
			b.WriteString(backtickIdentifiers(applyDescModifiers(p[last:m[0]])))
			b.WriteString(p[m[0]:m[1]])
			last = m[1]
		}
		b.WriteString(backtickIdentifiers(applyDescModifiers(p[last:])))
		return b.String()
	}

	if !preserveParagraphs {
		return collapse(s)
	}

	paragraphs := reParagraph.Split(s, -1)
	out := make([]string, 0, len(paragraphs))
	for _, p := range paragraphs {
		if c := collapse(p); c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, "\n\n")
}

// cutDeprecated splits a description with the "DEPRECATED:" prefix, e.g., "DEPRECATED: use index_patterns instead",
// into the title, which describes the field then, e.g., "(DEPRECATED) Maximum index count", and the deprecation message.
func cutDeprecated(description, title string) (string, string, bool) {
	msg, ok := strings.CutPrefix(description, "DEPRECATED:")
	if !ok {
		return description, "", false
	}

	msg = firstUpper(strings.TrimSpace(msg))
	if !strings.HasSuffix(msg, ".") {
		msg += "."
	}
	title = strings.TrimSpace(strings.TrimPrefix(title, "(DEPRECATED)"))

	// The docs render the message as is
	return or(title, msg), normalizeDescription(msg, false), true
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
	description := normalizeDescription(item.Description, item.IsRoot())

	b := userconfig.Desc(description)
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
	return nil
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
