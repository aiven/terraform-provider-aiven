package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

const (
	userConfigSuffix       = "_user_config"
	userConfigSchemaSuffix = "UserConfig"
	userConfigNoReset      = "**Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later."
	userConfigFieldsFunc   = "userConfigFields"
	expandUserConfig       = "ExpandUserConfig"
	flattenUserConfig      = "FlattenUserConfig"
)

// userConfigSource is a "oneOf" field found in an operation's request or response.
type userConfigSource struct {
	operation *Operation
	name      string    // The API field, e.g. "user_config"
	schema    *OASchema // The "oneOf" schema
	appearsIn AppearsIn
}

// addUserConfigs adds a "<kind>_user_config" block per "oneOf" schema, see Definition.UserConfigType.
// Runs once all operations are read, so the type attribute has its enum to match the schemas against.
func addUserConfigs(scope *Scope, root *Item) error {
	def := scope.Definition
	if def.UserConfigType == "" {
		return nil
	}

	typ, ok := root.Properties[def.UserConfigType]
	if !ok {
		return fmt.Errorf("userConfigType: %q not found", def.UserConfigType)
	}

	kinds := lo.Map(typ.Enum, func(v any, _ int) string { return fmt.Sprint(v) })
	if len(kinds) == 0 {
		return fmt.Errorf("userConfigType: %q has no enum", def.UserConfigType)
	}

	if len(scope.userConfigs) == 0 {
		return errors.New("userConfigType: no root field with a oneOf schema found")
	}

	for _, src := range scope.userConfigs {
		if def.userConfigField != "" && def.userConfigField != src.name {
			return fmt.Errorf("userConfigType: multiple oneOf fields found: %q and %q", def.userConfigField, src.name)
		}
		def.userConfigField = src.name

		var prefix string
		seen := make(map[string]string)
		for _, ref := range src.schema.OneOf {
			s, err := scope.OpenAPI.getComponentRef(ref.Ref)
			if err != nil {
				return fmt.Errorf("userConfigType: %w", err)
			}

			kind, p, err := userConfigKind(s.name, kinds)
			if err != nil {
				return fmt.Errorf("userConfigType: %w", err)
			}

			// A schema named differently from the others likely matched a wrong value,
			// e.g. "IntegrationEndpointExternalRedisUserConfig" matches "redis" when there is no "external_redis".
			if prefix != "" && p != prefix {
				return fmt.Errorf("userConfigType: schema %q doesn't have the prefix %q of the other schemas", s.name, prefix)
			}
			prefix = p

			if other, ok := seen[kind]; ok {
				return fmt.Errorf("userConfigType: schemas %q and %q match the same value %q", other, s.name, kind)
			}
			seen[kind] = s.name

			// SDKv2 has no blocks without options
			if len(s.Properties) == 0 {
				continue
			}

			desc := pascalCase(kind) + " user configurable settings. " + userConfigNoReset
			item := newItem(scope, root, kind+userConfigSuffix, &OASchema{Type: SchemaTypeObject, Description: desc}, false, src.appearsIn)
			item.UserConfigKind = kind

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
			return fmt.Errorf("userConfigType: no %q schema matches: %s", src.name, strings.Join(kinds, ", "))
		}
	}
	return nil
}

// userConfigKind returns the user config type of the "oneOf" schema, and the rest of the schema name.
// The longest match wins: "IntegrationEndpointExternalPrometheusUserConfig" is "external_prometheus", not "prometheus".
func userConfigKind(schemaName string, kinds []string) (string, string, error) {
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
func keepUserConfigState(item *Item) {
	for _, v := range item.Properties {
		if !v.InUserConfig() {
			continue
		}

		if !v.IsNested() && v.Computed && !v.Required && v.OverrideUseStateForUnknown == nil {
			v.OverrideUseStateForUnknown = new(true)
		}

		keepUserConfigState(v)
		if v.Items != nil {
			keepUserConfigState(v.Items)
		}
	}
}

// genUserConfigFields renders the function that maps the user config types to the user config blocks.
func genUserConfigFields(def *Definition, root *Item) jen.Code {
	if def.UserConfigType == "" {
		return nil
	}

	fields := make(jen.Dict)
	for _, v := range root.Properties {
		if v.UserConfigKind != "" {
			fields[jen.Lit(v.UserConfigKind)] = jen.Lit(v.Name)
		}
	}

	m := jen.Map(jen.String()).String()
	return jen.Commentf("%s maps %s values to the user config blocks", userConfigFieldsFunc, def.UserConfigType).
		Line().
		Func().Id(userConfigFieldsFunc).Params().Add(m.Clone()).
		Block(jen.Return(m.Clone().Values(fields)))
}

// userConfigArg renders the user config modifier when the user config blocks appear in the operation's
// request (Expand) or response (Flatten).
func userConfigArg(def *Definition, root *Item, operation *Operation, appearsIn AppearsIn, modifier string) jen.Code {
	if def.UserConfigType == "" {
		return nil
	}

	found := lo.SomeBy(lo.Values(root.Properties), func(v *Item) bool {
		return v.UserConfigKind != "" && v.AppearsIn.Contains(appearsIn)
	})
	if !found {
		return nil
	}

	args := []jen.Code{
		jen.Lit(def.UserConfigType),
		jen.Lit(def.userConfigField),
		jen.Id(userConfigFieldsFunc).Call(),
	}

	if modifier == expandUserConfig {
		required := operation.Request != nil && slices.Contains(operation.Request.Required, def.userConfigField)
		args = append(args, jen.Lit(required))
	}
	return jen.Qual(adapterPackage, modifier).Call(args...)
}
