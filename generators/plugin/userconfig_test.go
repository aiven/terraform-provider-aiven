package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserConfigType(t *testing.T) {
	t.Parallel()

	t.Run("longest suffix wins", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigType(
			"IntegrationEndpointExternalPrometheusUserConfig",
			[]string{"prometheus", "external_prometheus"},
		)
		require.NoError(t, err)
		assert.Equal(t, "external_prometheus", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("plain match", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigType(
			"IntegrationEndpointDatadogUserConfig",
			[]string{"datadog"},
		)
		require.NoError(t, err)
		assert.Equal(t, "datadog", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("kinds with digits", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigType(
			"IntegrationEndpointExternalAwsS3UserConfig",
			[]string{"external_aws_s3"},
		)
		require.NoError(t, err)
		assert.Equal(t, "external_aws_s3", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()
		_, _, err := userConfigType("IntegrationEndpointFooUserConfig", []string{"datadog"})
		require.Error(t, err)
		assert.ErrorContains(t, err, `matches no user config type`)
	})
}

func TestPascalCase(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ExternalAwsS3", pascalCase("external_aws_s3"))
	assert.Equal(t, "Datadog", pascalCase("datadog"))
}

func TestUserConfigDescription(t *testing.T) {
	t.Parallel()

	cases := map[string]*OASchema{
		"Tags with prefix `aiven-` are reserved.": {Description: "Tags with prefix 'aiven-' are reserved."},
		"Values: `delete`, `compact`.":            {Description: "Values: 'delete', 'compact'."},
		"Don't touch apostrophes.":                {Description: "Don't touch apostrophes."},
		"A longer `title`":                        {Description: "Short", Title: "A longer 'title'"},
		"The `grafana_stat_total_*` metrics.":     {Description: "The grafana_stat_total_* metrics."},
		"Matches `public-*`.":                     {Description: "Matches public-*."},
	}
	for want, s := range cases {
		assert.Equal(t, want, userConfigDescription(s))
	}
}

func TestKeepUserConfigState(t *testing.T) {
	t.Parallel()

	falseOverride := new(false)
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"outside": {Name: "outside", Type: SchemaTypeString},
		},
	}
	variant := &Item{
		Name:           "datadog_user_config",
		Type:           SchemaTypeObject,
		UserConfigType: "datadog",
		Parent:         root,
		Properties: map[string]*Item{
			"computed_optional": {
				Name:     "computed_optional",
				Type:     SchemaTypeString,
				Computed: true,
			},
			"required": {
				Name:     "required",
				Type:     SchemaTypeString,
				Required: true,
			},
			"plain": {
				Name: "plain",
				Type: SchemaTypeString,
			},
			"explicit_false": {
				Name:                       "explicit_false",
				Type:                       SchemaTypeString,
				Computed:                   true,
				OverrideUseStateForUnknown: falseOverride,
			},
			"ip_filter_string": {
				Name:          "ip_filter_string",
				Type:          SchemaTypeString,
				Computed:      true,
				ConflictsWith: []string{"ip_filter"},
			},
		},
	}
	root.Properties["datadog_user_config"] = variant
	for _, item := range variant.Properties {
		item.Parent = variant
	}
	root.Properties["outside"].Parent = root

	tables := &Item{Name: "tables", Type: SchemaTypeArrayOrdered, Computed: true, Parent: variant}
	tables.Items = &Item{Type: SchemaTypeObject, Parent: tables}
	tableOption := &Item{Name: "auto_offset_reset", Type: SchemaTypeString, Computed: true, Parent: tables.Items}
	tables.Items.Properties = map[string]*Item{"auto_offset_reset": tableOption}
	variant.Properties["tables"] = tables

	keepUserConfigState(root)

	require.NotNil(t, variant.Properties["computed_optional"].OverrideUseStateForUnknown)
	assert.True(t, *variant.Properties["computed_optional"].OverrideUseStateForUnknown)
	assert.Nil(t, variant.Properties["required"].OverrideUseStateForUnknown)
	assert.Nil(t, variant.Properties["plain"].OverrideUseStateForUnknown)
	assert.Same(t, falseOverride, variant.Properties["explicit_false"].OverrideUseStateForUnknown)
	assert.Nil(t, variant.Properties["ip_filter_string"].OverrideUseStateForUnknown)
	assert.Nil(t, root.Properties["outside"].OverrideUseStateForUnknown)
	assert.Nil(t, tableOption.OverrideUseStateForUnknown, "list items are matched by index")
}

func TestAddUserConfigs(t *testing.T) {
	t.Parallel()

	openAPI := &OpenAPIDoc{}
	openAPI.Components.Schemas = map[string]*OASchema{
		"IntegrationEndpointDatadogUserConfig": {
			Type: SchemaTypeObject,
			Properties: map[string]*OASchema{
				"site":       {Type: SchemaTypeString},
				"access_key": {Type: SchemaTypeString},
				"api_key":    {Type: SchemaTypeString, SensitiveExt: true},
				"tags":       {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
			},
		},
		"IntegrationEndpointRsyslogUserConfig": {
			Type: SchemaTypeObject,
			Properties: map[string]*OASchema{
				"server": {Type: SchemaTypeString},
			},
		},
		"IntegrationEndpointJolokiaUserConfig": {
			Type: SchemaTypeObject,
		},
	}
	scope := &Scope{
		OpenAPI: openAPI,
		Definition: &Definition{
			UserConfigDiscriminator: "endpoint_type",
		},
		userConfigs: []*userConfigSource{{
			operation: &Operation{Type: OperationCreate},
			name:      "user_config",
			schema: &OASchema{
				OneOf: []*OASchema{
					{Ref: "#/components/schemas/IntegrationEndpointDatadogUserConfig"},
					{Ref: "#/components/schemas/IntegrationEndpointRsyslogUserConfig"},
					{Ref: "#/components/schemas/IntegrationEndpointJolokiaUserConfig"},
				},
			},
			appearsIn: RequestBody,
		}},
	}
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"endpoint_type": {
				Name: "endpoint_type",
				Enum: []any{"datadog", "jolokia", "rsyslog"},
			},
		},
	}

	require.NoError(t, addUserConfigs(scope, root))

	datadog := root.Properties["datadog_user_config"]
	require.NotNil(t, datadog)
	assert.Equal(t, "datadog", datadog.UserConfigType)
	assert.True(t, datadog.Computed)
	require.Contains(t, datadog.Properties, "site")
	assert.Equal(t, SchemaTypeArrayOrdered, datadog.Properties["tags"].Type, "arrays are lists")
	assert.False(t, datadog.Properties["access_key"].Sensitive, "only x-_secure is sensitive")
	assert.True(t, datadog.Properties["api_key"].Sensitive)

	rsyslog := root.Properties["rsyslog_user_config"]
	require.NotNil(t, rsyslog)
	assert.Equal(t, "rsyslog", rsyslog.UserConfigType)
	assert.True(t, rsyslog.Computed)
	require.Contains(t, rsyslog.Properties, "server")

	assert.NotContains(t, root.Properties, "jolokia_user_config", "a schema without options has no block")
}

func TestAddUserConfigsFieldName(t *testing.T) {
	t.Parallel()

	scope := &Scope{
		Definition:  &Definition{UserConfigDiscriminator: "endpoint_type"},
		userConfigs: []*userConfigSource{{name: "config", schema: &OASchema{}}},
	}
	root := &Item{Properties: map[string]*Item{"endpoint_type": {Enum: []any{"datadog"}}}}
	assert.EqualError(t, addUserConfigs(scope, root), `userConfigDiscriminator: oneOf field "config" is not "user_config"`)
}

func TestAddUserConfigsFixedType(t *testing.T) {
	t.Parallel()

	newScope := func(kind string) *Scope {
		openAPI := &OpenAPIDoc{}
		openAPI.Components.Schemas = map[string]*OASchema{
			"ServicePgUserConfig": {
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"ip_filter": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							AnyOf:      []*OASchema{{Type: SchemaTypeString}, {Type: SchemaTypeObject}},
							Properties: map[string]*OASchema{"network": {Type: SchemaTypeString}},
						},
					},
					"pg_version": {Type: SchemaTypeString, Enum: []any{"17", "18"}},
					"pg": {
						Type: SchemaTypeObject,
						Properties: map[string]*OASchema{
							"pg_partman_bgw.role": {Type: SchemaTypeString},
							"log_line_prefix":     {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
							"access_key":          {Type: SchemaTypeString},
							"password":            {Type: SchemaTypeString, SensitiveExt: true},
						},
					},
				},
			},
			"ServiceKafkaUserConfig": {
				Type:       SchemaTypeObject,
				Properties: map[string]*OASchema{"kafka_rest": {Type: SchemaTypeBoolean}},
			},
		}
		return &Scope{
			OpenAPI: openAPI,
			Definition: &Definition{
				UserConfigDiscriminator: "service_type",
				Operations: Operations{{
					ID:          "ServiceCreate",
					Type:        OperationCreate,
					FixedValues: map[string]string{"service_type": kind},
				}},
			},
			userConfigs: []*userConfigSource{{
				operation: &Operation{Type: OperationCreate},
				name:      "user_config",
				schema: &OASchema{OneOf: []*OASchema{
					{Ref: "#/components/schemas/ServiceKafkaUserConfig"},
					{Ref: "#/components/schemas/ServicePgUserConfig"},
				}},
				appearsIn: RequestBody,
			}},
		}
	}
	newRoot := func() *Item {
		return &Item{
			Name:       "root",
			Type:       SchemaTypeObject,
			Properties: map[string]*Item{"service_type": {Name: "service_type", JSONName: "service_type", Type: SchemaTypeString}},
		}
	}

	t.Run("only the fixed type gets a block", func(t *testing.T) {
		t.Parallel()
		root := newRoot()
		require.NoError(t, addUserConfigs(newScope("pg"), root))
		assert.NotContains(t, root.Properties, "kafka_user_config")

		pg := root.Properties["pg_user_config"]
		require.NotNil(t, pg)

		for _, k := range []string{"ip_filter", "ip_filter_string", "ip_filter_object"} {
			require.Contains(t, pg.Properties, k)
			assert.Equal(t, SchemaTypeArray, pg.Properties[k].Type, "%s is a set", k)
			assert.Equal(t, "ip_filter", pg.Properties[k].JSONName)
		}
		assert.Equal(t, "Deprecated. Use `ip_filter_string` instead.", pg.Properties["ip_filter"].DeprecationMessage)
		assert.Equal(t, SchemaTypeString, pg.Properties["ip_filter"].Items.Type)
		assert.Equal(t, SchemaTypeString, pg.Properties["ip_filter_string"].Items.Type)
		assert.Equal(t, SchemaTypeObject, pg.Properties["ip_filter_object"].Items.Type)
		assert.Contains(t, pg.Properties["ip_filter_object"].Items.Properties, "network")

		version := pg.Properties["pg_version"]
		assert.Empty(t, version.Enum, "a version isn't validated")
		assert.Equal(t, []any{"17", "18"}, version.DocEnum)

		nested := pg.Properties["pg"]
		require.Contains(t, nested.Properties, "pg_partman_bgw__dot__role")
		assert.Equal(t, "pg_partman_bgw.role", nested.Properties["pg_partman_bgw__dot__role"].JSONName)
		assert.Equal(t, SchemaTypeArrayOrdered, nested.Properties["log_line_prefix"].Type, "other arrays are lists")
		assert.False(t, nested.Properties["access_key"].Sensitive, "only x-_secure is sensitive")
		assert.True(t, nested.Properties["password"].Sensitive)
	})

	t.Run("no schema matches", func(t *testing.T) {
		t.Parallel()
		err := addUserConfigs(newScope("mysql"), newRoot())
		require.ErrorContains(t, err, `no "user_config" schema matches: mysql`)
	})
}

func TestUserConfigProperties(t *testing.T) {
	t.Parallel()

	untyped := &OASchema{OneOf: []*OASchema{{Type: SchemaTypeObject}, {Type: SchemaTypeObject}}}
	properties := map[string]*OASchema{
		"ratio":   {AnyOf: []*OASchema{{Type: SchemaTypeInteger}, {Type: SchemaTypeNumber}}},
		"timeout": {Type: SchemaTypeInteger, AnyOf: []*OASchema{{}, {}}},
		"keystores": {Type: SchemaTypeArray, Items: &OASchema{
			Type:       SchemaTypeObject,
			Required:   []string{"settings"},
			Properties: map[string]*OASchema{"name": {Type: SchemaTypeString}, "settings": untyped},
		}},
		"optional": {Type: SchemaTypeObject, Properties: map[string]*OASchema{"settings": untyped}},
		"settings": untyped,
	}

	result := userConfigProperties(properties)
	assert.Equal(t, SchemaTypeNumber, result["ratio"].Type)
	assert.Nil(t, result["ratio"].AnyOf)
	assert.Empty(t, properties["ratio"].Type, "the original is intact")
	assert.Equal(t, SchemaTypeInteger, result["timeout"].Type)
	assert.NotContains(t, result, "settings")
	assert.NotContains(t, result, "keystores", "requires an untyped option")
	assert.Contains(t, result, "optional", "the untyped option is optional")
}

func TestSplitMultiTypeArrays(t *testing.T) {
	t.Parallel()

	single := &OASchema{Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}}
	mixed := &OASchema{
		Type: SchemaTypeArray,
		Items: &OASchema{
			AnyOf:      []*OASchema{{Type: SchemaTypeObject}, {Type: SchemaTypeString}},
			MaxLength:  43,
			Properties: map[string]*OASchema{"network": {Type: SchemaTypeString}},
			Required:   []string{"network"},
		},
	}
	got := splitMultiTypeArrays(map[string]*OASchema{"single": single, "mixed": mixed})

	assert.Len(t, got, 4)
	assert.Same(t, single, got["single"])

	str := got["mixed_string"]
	assert.Equal(t, "mixed", str.jsonName)
	assert.Equal(t, SchemaTypeString, str.Items.Type)
	assert.Equal(t, 43, str.Items.MaxLength)
	assert.Empty(t, str.Items.Properties)

	obj := got["mixed_object"]
	assert.Equal(t, SchemaTypeObject, obj.Items.Type)
	assert.Zero(t, obj.Items.MaxLength)
	assert.Equal(t, []string{"network"}, obj.Items.Required)

	orig := got["mixed"]
	assert.Equal(t, SchemaTypeString, orig.Items.Type, "scalars go first")
	assert.Equal(t, "Deprecated. Use `mixed_string` instead.", orig.deprecation)
	assert.Empty(t, mixed.Items.Type, "the original schema is intact")
}
