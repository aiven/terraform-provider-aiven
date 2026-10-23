package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserConfigKind(t *testing.T) {
	t.Parallel()

	t.Run("longest suffix wins", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigKind(
			"IntegrationEndpointExternalPrometheusUserConfig",
			[]string{"prometheus", "external_prometheus"},
		)
		require.NoError(t, err)
		assert.Equal(t, "external_prometheus", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("plain match", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigKind(
			"IntegrationEndpointDatadogUserConfig",
			[]string{"datadog"},
		)
		require.NoError(t, err)
		assert.Equal(t, "datadog", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("kinds with digits", func(t *testing.T) {
		t.Parallel()
		kind, prefix, err := userConfigKind(
			"IntegrationEndpointExternalAwsS3UserConfig",
			[]string{"external_aws_s3"},
		)
		require.NoError(t, err)
		assert.Equal(t, "external_aws_s3", kind)
		assert.Equal(t, "IntegrationEndpoint", prefix)
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()
		_, _, err := userConfigKind("IntegrationEndpointFooUserConfig", []string{"datadog"})
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
		UserConfigKind: "datadog",
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
		},
	}
	root.Properties["datadog_user_config"] = variant
	for _, item := range variant.Properties {
		item.Parent = variant
	}
	root.Properties["outside"].Parent = root

	keepUserConfigState(root)

	require.NotNil(t, variant.Properties["computed_optional"].OverrideUseStateForUnknown)
	assert.True(t, *variant.Properties["computed_optional"].OverrideUseStateForUnknown)
	assert.Nil(t, variant.Properties["required"].OverrideUseStateForUnknown)
	assert.Nil(t, variant.Properties["plain"].OverrideUseStateForUnknown)
	assert.Same(t, falseOverride, variant.Properties["explicit_false"].OverrideUseStateForUnknown)
	assert.Nil(t, root.Properties["outside"].OverrideUseStateForUnknown)
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
			UserConfigType: "endpoint_type",
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
	assert.Equal(t, "user_config", scope.Definition.userConfigField)

	datadog := root.Properties["datadog_user_config"]
	require.NotNil(t, datadog)
	assert.Equal(t, "datadog", datadog.UserConfigKind)
	assert.True(t, datadog.Computed)
	require.Contains(t, datadog.Properties, "site")
	assert.Equal(t, SchemaTypeArrayOrdered, datadog.Properties["tags"].Type, "arrays are lists")
	assert.False(t, datadog.Properties["access_key"].Sensitive, "only x-_secure is sensitive")
	assert.True(t, datadog.Properties["api_key"].Sensitive)

	rsyslog := root.Properties["rsyslog_user_config"]
	require.NotNil(t, rsyslog)
	assert.Equal(t, "rsyslog", rsyslog.UserConfigKind)
	assert.True(t, rsyslog.Computed)
	require.Contains(t, rsyslog.Properties, "server")

	assert.NotContains(t, root.Properties, "jolokia_user_config", "a schema without options has no block")
}
