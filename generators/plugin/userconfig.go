package main

import (
	"errors"
	"fmt"
	"regexp"
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

	kinds := lo.Map(typ.Enum, func(v any, _ int) string { return fmt.Sprint(v) })
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
			if err != nil {
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

// reUserConfigCode matches a single-quoted token without whitespace, e.g. 'aiven-'.
var reUserConfigCode = regexp.MustCompile(`'([^'\s]+)'`)

// userConfigDescription prefers the longer of the description and the title,
// and wraps single-quoted tokens in backticks, as SDKv2 user configs do.
func userConfigDescription(s *OASchema) string {
	d := s.Description
	if len(d) < len(s.Title) {
		d = s.Title
	}
	return reUserConfigCode.ReplaceAllString(d, "`$1`")
}

// keepUserConfigState plans the user config options missing in the config with their state values.
// Without it, they are planned as unknown on every update.
// Skips the list items: the framework matches them by index, so removing one would pass its state to the next one.
func keepUserConfigState(item *Item) {
	for _, v := range item.Properties {
		if !v.InUserConfig() {
			continue
		}

		if !v.IsNested() && v.Computed && !v.Required && v.OverrideUseStateForUnknown == nil {
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
