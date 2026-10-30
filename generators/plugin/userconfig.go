package main

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

const (
	userConfigField        = "user_config" // The API field the adapter's modifiers expect
	userConfigSuffix       = "_user_config"
	userConfigSchemaSuffix = "UserConfig"
	userConfigNoReset      = "**Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later."
	expandUserConfig       = "ExpandUserConfig"
	flattenUserConfig      = "FlattenUserConfig"
	validateUserConfig     = "ValidateUserConfig"
)

// userConfigSource is a "oneOf" field found in an operation's request or response.
type userConfigSource struct {
	operation *Operation
	name      string    // The API field, e.g. "user_config"
	schema    *OASchema // The "oneOf" schema
	appearsIn AppearsIn
}

// addUserConfigs adds a "<type>_user_config" block per "oneOf" schema, see Definition.UserConfigDiscriminator.
// Runs once all operations are read, so the type attribute has its enum to match the schemas against.
func addUserConfigs(scope *Scope, root *Item) error {
	def := scope.Definition
	if def.UserConfigDiscriminator == "" {
		return nil
	}

	typ, ok := root.Properties[def.UserConfigDiscriminator]
	if !ok {
		return fmt.Errorf("userConfigDiscriminator: %q not found", def.UserConfigDiscriminator)
	}

	// A fixed type gets its block only, e.g., a service resource has a single type.
	enum := typ.Enum
	fixed, isFixed := def.FixedValue(typ.JSONName)
	if isFixed {
		enum = []any{fixed}
	}

	kinds := lo.Map(enum, func(v any, _ int) string { return fmt.Sprint(v) })
	if len(kinds) == 0 {
		return fmt.Errorf("userConfigDiscriminator: %q has no enum", def.UserConfigDiscriminator)
	}

	if len(scope.userConfigs) == 0 {
		return errors.New("userConfigDiscriminator: no root field with a oneOf schema found")
	}

	for _, src := range scope.userConfigs {
		if src.name != userConfigField {
			return fmt.Errorf("userConfigDiscriminator: oneOf field %q is not %q", src.name, userConfigField)
		}

		var prefix string
		seen := make(map[string]string)
		for _, ref := range src.schema.OneOf {
			s, err := scope.OpenAPI.getComponentRef(ref.Ref)
			if err != nil {
				return fmt.Errorf("userConfigDiscriminator: %w", err)
			}

			kind, p, err := userConfigType(s.name, kinds)
			switch {
			case err != nil && isFixed:
				continue
			case err != nil:
				return fmt.Errorf("userConfigDiscriminator: %w", err)
			}

			// A schema named differently from the others likely matched a wrong value,
			// e.g. "IntegrationEndpointExternalRedisUserConfig" matches "redis" when there is no "external_redis".
			if prefix != "" && p != prefix {
				return fmt.Errorf("userConfigDiscriminator: schema %q doesn't have the prefix %q of the other schemas", s.name, prefix)
			}
			prefix = p

			if other, ok := seen[kind]; ok {
				return fmt.Errorf("userConfigDiscriminator: schemas %q and %q match the same value %q", other, s.name, kind)
			}
			seen[kind] = s.name

			// SDKv2 has no blocks without options
			if len(s.Properties) == 0 {
				continue
			}

			desc := pascalCase(kind) + " user configurable settings. " + userConfigNoReset
			item := newItem(scope, root, kind+userConfigSuffix, &OASchema{Type: SchemaTypeObject, Description: desc}, false, src.appearsIn)
			item.UserConfigType = kind

			// The API keeps the options removed from the config.
			// Computed blocks keep their state values, see keepUserConfigState.
			item.Computed = true

			err = fromSchema(scope, src.operation, item, s, src.appearsIn)
			if err != nil {
				return fmt.Errorf("userConfig %q: %w", kind, err)
			}

			err = addProperty(scope, root, item)
			if err != nil {
				return fmt.Errorf("userConfig %q: %w", kind, err)
			}
		}

		if len(seen) == 0 {
			return fmt.Errorf("userConfigDiscriminator: no %q schema matches: %s", src.name, strings.Join(kinds, ", "))
		}
	}
	return nil
}

// userConfigType returns the user config type of the "oneOf" schema, and the rest of the schema name.
// The longest match wins: "IntegrationEndpointExternalPrometheusUserConfig" is "external_prometheus", not "prometheus".
func userConfigType(schemaName string, kinds []string) (string, string, error) {
	var kind, suffix string
	for _, k := range kinds {
		s := pascalCase(k) + userConfigSchemaSuffix
		if strings.HasSuffix(schemaName, s) && len(s) > len(suffix) {
			kind, suffix = k, s
		}
	}

	if kind == "" {
		return "", "", fmt.Errorf("schema %q matches no user config type: %s", schemaName, strings.Join(kinds, ", "))
	}
	return kind, strings.TrimSuffix(schemaName, suffix), nil
}

// pascalCase turns "external_aws_s3" into "ExternalAwsS3".
func pascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		parts[i] = firstUpper(p)
	}
	return strings.Join(parts, "")
}

var (
	// reUserConfigCode matches a single-quoted token without whitespace, e.g. 'aiven-'.
	reUserConfigCode = regexp.MustCompile(`'([^'\s]+)'`)
	// reUserConfigGlob matches an unquoted wildcard name, e.g. public-*, which Markdown parses as emphasis.
	reUserConfigGlob = regexp.MustCompile(`(^|\s)([\w-]+\*[\w*-]*)`)
)

// userConfigDescription prefers the longer of the description and the title,
// and wraps single-quoted tokens and wildcard names in backticks, as SDKv2 user configs do.
func userConfigDescription(s *OASchema) string {
	d := s.Description
	if len(d) < len(s.Title) {
		d = s.Title
	}
	d = reUserConfigCode.ReplaceAllString(d, "`$1`")
	return reUserConfigGlob.ReplaceAllString(d, "$1`$2`")
}

// userConfigProperties types the options as SDKv2 does, see splitMultiTypeArrays.
// An integer or number is a number. The options without a type are skipped,
// and so are the objects that require them.
func userConfigProperties(properties map[string]*OASchema) map[string]*OASchema {
	result := splitMultiTypeArrays(properties)
	for name, s := range result {
		types := anyOfTypes(s)
		if s.Type == "" && len(types) > 0 && !slices.ContainsFunc(types, func(t SchemaType) bool {
			return t != SchemaTypeInteger && t != SchemaTypeNumber
		}) {
			number := *s
			number.Type, number.AnyOf = SchemaTypeNumber, nil
			result[name] = &number
		}

		if !isTyped(result[name]) {
			delete(result, name)
		}
	}
	return result
}

// isTyped tells if the option and the options it requires have types.
// Arrays of mixed item types are typed: they are split.
func isTyped(s *OASchema) bool {
	switch {
	case s.Ref != "":
		return true
	case s.Type == SchemaTypeArray && s.Items != nil:
		return len(anyOfTypes(s.Items)) > 1 || isTyped(s.Items)
	case s.Type == SchemaTypeObject:
		return !slices.ContainsFunc(s.Required, func(k string) bool {
			p, ok := s.Properties[k]
			return ok && !isTyped(p)
		})
	}
	return s.Type != ""
}

// anyOfTypes returns the unique types of the "anyOf" schemas.
func anyOfTypes(s *OASchema) []SchemaType {
	return lo.Uniq(lo.FilterMap(s.AnyOf, func(v *OASchema, _ int) (SchemaType, bool) {
		return v.Type, v.Type != ""
	}))
}

// splitMultiTypeArrays splits an array of mixed item types into a field per type, as SDKv2 does:
// "ip_filter" with string and object items gets "ip_filter_string" and "ip_filter_object",
// and keeps the first type, deprecated. A field gains new types over time, so scalars go first.
// The fields set the same option, so they conflict with each other.
func splitMultiTypeArrays(properties map[string]*OASchema) map[string]*OASchema {
	result := make(map[string]*OASchema, len(properties))
	for name, s := range properties {
		result[name] = s
		if s.Type != SchemaTypeArray || s.Items == nil {
			continue
		}

		types := anyOfTypes(s.Items)
		if len(types) < 2 {
			continue
		}

		slices.SortFunc(types, func(a, b SchemaType) int {
			isObject := func(t SchemaType) int { return lo.Ternary(t == SchemaTypeObject, 1, 0) }
			return cmp.Or(isObject(a)-isObject(b), cmp.Compare(a, b))
		})

		for _, t := range types {
			items := *s.Items
			items.Type, items.AnyOf = t, nil
			if t == SchemaTypeObject {
				items.MaxLength, items.Pattern, items.Format = 0, "", ""
			} else {
				items.Properties, items.Required, items.AdditionalProperties = nil, nil, nil
			}

			field := *s
			field.Items = &items
			field.jsonName = name
			result[name+"_"+string(t)] = &field
		}

		field := *result[name+"_"+string(types[0])]
		field.deprecation = fmt.Sprintf("Deprecated. Use `%s_%s` instead.", name, types[0])
		result[name] = &field

		names := []string{name}
		for _, t := range types {
			names = append(names, name+"_"+string(t))
		}
		for _, n := range names {
			result[n].conflictsWith = lo.Without(names, n)
		}
	}
	return result
}

// isUserConfigSet tells if the user config field is a set: SDKv2 has sets for the IP filters only.
func isUserConfigSet(item *Item) bool {
	return item.Parent.UserConfigType != "" && item.JSONName == "ip_filter"
}

// isLegacyUserConfigName tells if SDKv2 names the option with "__dot__" instead of "_" for dots.
func isLegacyUserConfigName(name string) bool {
	return strings.HasPrefix(name, "pg_partman_bgw") || strings.HasPrefix(name, "pg_stat_")
}

// keepUserConfigState plans the user config options missing in the config with their state values.
// Without it, they are planned as unknown on every update.
// Skips the list items: the framework matches them by index, so removing one would pass its state to the next one.
// Skips the options of the same API field too: switching ip_filter_string to ip_filter_object would send both.
func keepUserConfigState(item *Item) {
	for _, v := range item.Properties {
		if !v.InUserConfig() {
			continue
		}

		if !v.IsNested() && v.Computed && !v.Required && len(v.ConflictsWith) == 0 && v.OverrideUseStateForUnknown == nil {
			v.OverrideUseStateForUnknown = new(true)
		}
		keepUserConfigState(v)
	}
}

// userConfigArg renders the user config modifier when the user config blocks appear in the operation's
// request (Expand) or response (Flatten).
func userConfigArg(def *Definition, root *Item, appearsIn AppearsIn, modifier string) jen.Code {
	if def.UserConfigDiscriminator == "" {
		return nil
	}

	found := lo.SomeBy(lo.Values(root.Properties), func(v *Item) bool {
		return v.UserConfigType != "" && v.AppearsIn.Contains(appearsIn)
	})
	if !found {
		return nil
	}

	return jen.Qual(adapterPackage, modifier).Call(jen.Lit(def.UserConfigDiscriminator))
}
