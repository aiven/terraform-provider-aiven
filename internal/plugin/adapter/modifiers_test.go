// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandUserConfig(t *testing.T) {
	t.Parallel()

	mod := ExpandUserConfig("endpoint_type")
	d, err := NewResourceData(userConfigTestSchema(), nil)
	require.NoError(t, err)

	t.Run("selected block becomes user_config", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": map[string]any{"site": "datadoghq.com"},
			"rsyslog_user_config": map[string]any{"server": "syslog.example.com"},
			"display_name":        "foo",
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{"site": "datadoghq.com"}, dto["user_config"])
		assert.NotContains(t, dto, "datadog_user_config")
		assert.NotContains(t, dto, "rsyslog_user_config")
		assert.Equal(t, "foo", dto["display_name"])
	})

	t.Run("missing selected block sends empty map", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type":       "datadog",
			"rsyslog_user_config": map[string]any{"server": "syslog.example.com"},
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{}, dto["user_config"])
	})

	t.Run("empty slice block sends empty map", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{},
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{}, dto["user_config"])
	})

	t.Run("unset nested objects are not sent", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type": "datadog",
			"datadog_user_config": map[string]any{
				"site":       "datadoghq.com",
				"opensearch": []any{},
				"tables": []any{
					map[string]any{"name": "foo", "settings": []any{}, "topics": []any{}},
					map[string]any{"name": "bar", "settings": map[string]any{"max_block_size": 1}},
				},
			},
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{
			"site": "datadoghq.com",
			"tables": []any{
				map[string]any{"name": "foo", "topics": []any{}},
				map[string]any{"name": "bar", "settings": map[string]any{"max_block_size": 1}},
			},
		}, dto["user_config"])
	})
}

func TestValidateUserConfig(t *testing.T) {
	t.Parallel()

	sch := userConfigTestSchema()
	sch.Properties["external_postgresql"] = userConfigTestSchema().Properties["rsyslog_user_config"]
	validate := ValidateUserConfig("endpoint_type")
	block := []any{map[string]any{"server": "syslog.example.com"}}

	cases := []struct {
		name   string
		config map[string]any
		err    string
	}{
		{
			name:   "selected block",
			config: map[string]any{"endpoint_type": "rsyslog", "rsyslog_user_config": block, "datadog_user_config": []any{}},
		},
		{
			name:   "block of another type",
			config: map[string]any{"endpoint_type": "datadog", "rsyslog_user_config": block},
			err:    `rsyslog_user_config can't be set when endpoint_type is "datadog"`,
		},
		{
			name:   "legacy block",
			config: map[string]any{"endpoint_type": "external_postgresql", "external_postgresql": block},
		},
		{
			name:   "legacy block of another type",
			config: map[string]any{"endpoint_type": "datadog", "external_postgresql": block},
			err:    `external_postgresql can't be set when endpoint_type is "datadog"`,
		},
		{
			name:   "unknown type",
			config: map[string]any{"rsyslog_user_config": block},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, err := NewResourceData(sch, nil, WithTestPlan(tc.config))
			require.NoError(t, err)
			err = validate(t.Context(), nil, d)
			if tc.err == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tc.err)
		})
	}
}

func userConfigTestSchema() *Schema {
	object := func(props map[string]*Schema) *Schema {
		return &Schema{Type: SchemaTypeList, IsObject: true, Computed: true, Items: &Schema{Type: SchemaTypeObject, Properties: props}}
	}
	return &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id":            {Type: SchemaTypeString, Computed: true},
			"endpoint_type": {Type: SchemaTypeString},
			"datadog_user_config": object(map[string]*Schema{
				"site":            {Type: SchemaTypeString, Computed: true},
				"max_jmx_metrics": {Type: SchemaTypeInt, Computed: true},
				"datadog_tags":    {Type: SchemaTypeList, Computed: true, Items: &Schema{Type: SchemaTypeString}},
				"tables": {Type: SchemaTypeList, Computed: true, Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"name":              {Type: SchemaTypeString, Computed: true},
						"auto_offset_reset": {Type: SchemaTypeString, Computed: true},
						"settings": object(map[string]*Schema{
							"max_block_size": {Type: SchemaTypeInt, Computed: true},
						}),
						"topics": {Type: SchemaTypeList, Computed: true, Items: &Schema{
							Type: SchemaTypeObject,
							Properties: map[string]*Schema{
								"name":      {Type: SchemaTypeString, Computed: true},
								"partition": {Type: SchemaTypeInt, Computed: true},
							},
						}},
					},
				}},
				"opensearch": object(map[string]*Schema{
					"index_stats_enabled":  {Type: SchemaTypeBool, Computed: true},
					"pshard_stats_enabled": {Type: SchemaTypeBool, Computed: true},
				}),
			}),
			"rsyslog_user_config": object(map[string]*Schema{
				"server": {Type: SchemaTypeString, Computed: true},
			}),
		},
	}
}

func TestFlattenUserConfig(t *testing.T) {
	t.Parallel()

	mod := FlattenUserConfig("endpoint_type")
	dataSource, err := NewResourceData(userConfigTestSchema(), nil, WithIsDataSource())
	require.NoError(t, err)

	// The API returns the defaults too
	response := func() map[string]any {
		return map[string]any{
			"endpoint_type": "datadog",
			"user_config": map[string]any{
				"site":            "datadoghq.com",
				"max_jmx_metrics": 2000,
				"datadog_tags":    []any{"a", "b"},
				"opensearch":      map[string]any{"index_stats_enabled": true, "pshard_stats_enabled": false},
			},
		}
	}

	t.Run("data source keeps all options", func(t *testing.T) {
		t.Parallel()
		dto := response()
		want := dto["user_config"]
		require.NoError(t, mod(dataSource, dto))
		require.Equal(t, want, dto["datadog_user_config"])
		assert.NotContains(t, dto, "user_config")
	})

	t.Run("data source sets no block for empty user_config", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type": "datadog",
			"user_config":   map[string]any{},
		}
		require.NoError(t, mod(dataSource, dto))
		assert.NotContains(t, dto, "user_config")
		assert.NotContains(t, dto, "datadog_user_config")
	})

	t.Run("data source sets no block for nil user_config", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{"endpoint_type": "datadog"}
		require.NoError(t, mod(dataSource, dto))
		assert.NotContains(t, dto, "datadog_user_config")
	})

	t.Run("unknown kind sets no block", func(t *testing.T) {
		t.Parallel()
		dto := map[string]any{
			"endpoint_type": "unknown",
			"user_config":   map[string]any{"site": "x"},
		}
		require.NoError(t, mod(dataSource, dto))
		assert.NotContains(t, dto, "user_config")
		assert.NotContains(t, dto, "datadog_user_config")
		assert.NotContains(t, dto, "rsyslog_user_config")
	})

	t.Run("resource keeps options set in the plan", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type": "datadog",
			"datadog_user_config": []any{map[string]any{
				"site":         "datadoghq.com",
				"datadog_tags": []any{"a"},
				"opensearch":   []any{map[string]any{"index_stats_enabled": true}},
			}},
		}))
		require.NoError(t, err)

		dto := response()
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{
			"site":         "datadoghq.com",
			"datadog_tags": []any{"a", "b"},
			"opensearch":   map[string]any{"index_stats_enabled": true},
		}, dto["datadog_user_config"])
	})

	t.Run("resource keeps options set in the state on read", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestState(map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{map[string]any{"max_jmx_metrics": 1000}},
		}))
		require.NoError(t, err)

		dto := response()
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{"max_jmx_metrics": 2000}, dto["datadog_user_config"])
	})

	t.Run("resource sets no block when it is not set", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{},
		}))
		require.NoError(t, err)

		dto := response()
		require.NoError(t, mod(d, dto))
		assert.NotContains(t, dto, "datadog_user_config")
		assert.NotContains(t, dto, "user_config")
	})

	t.Run("resource keeps an empty block", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{map[string]any{}},
		}))
		require.NoError(t, err)

		dto := map[string]any{"endpoint_type": "datadog", "user_config": map[string]any{}}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{}, dto["datadog_user_config"])
	})

	t.Run("resource keeps the options set in each list object", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type": "datadog",
			"datadog_user_config": []any{map[string]any{
				"tables": []any{
					map[string]any{"name": "foo", "topics": []any{map[string]any{"name": "foo"}}},
					map[string]any{"name": "bar", "auto_offset_reset": "latest"},
				},
			}},
		}))
		require.NoError(t, err)

		dto := map[string]any{
			"endpoint_type": "datadog",
			"user_config": map[string]any{
				"site": "datadoghq.com",
				"tables": []any{
					map[string]any{
						"name":              "foo",
						"auto_offset_reset": "earliest",
						"topics":            []any{map[string]any{"name": "foo", "partition": 0}},
					},
					map[string]any{"name": "bar", "auto_offset_reset": "latest"},
				},
			},
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{
			"tables": []any{
				map[string]any{"name": "foo", "topics": []any{map[string]any{"name": "foo"}}},
				map[string]any{"name": "bar", "auto_offset_reset": "latest"},
			},
		}, dto["datadog_user_config"])
	})

	t.Run("resource drops the default list objects", func(t *testing.T) {
		t.Parallel()
		// A list block without blocks is planned as an empty list, not null
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{map[string]any{"site": "datadoghq.com", "tables": []any{}}},
		}))
		require.NoError(t, err)

		dto := map[string]any{
			"endpoint_type": "datadog",
			"user_config": map[string]any{
				"site":   "datadoghq.com",
				"tables": []any{map[string]any{"name": "default", "auto_offset_reset": "earliest"}},
			},
		}
		require.NoError(t, mod(d, dto))
		require.Equal(t, map[string]any{"site": "datadoghq.com", "tables": []any{}}, dto["datadog_user_config"])
	})

	t.Run("resource state mirrors the plan", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(userConfigTestSchema(), nil, WithTestPlan(map[string]any{
			"endpoint_type":       "datadog",
			"datadog_user_config": []any{map[string]any{"site": "datadoghq.com"}},
		}))
		require.NoError(t, err)

		require.NoError(t, d.Flatten(response(), mod))
		assert.Equal(t, []any{map[string]any{"site": "datadoghq.com"}}, d.Get("datadog_user_config"))
		_, ok := d.GetOk("rsyslog_user_config")
		assert.False(t, ok)
	})
}

func TestLegacyUserConfig(t *testing.T) {
	t.Parallel()

	sch := userConfigTestSchema()
	sch.Properties["external_postgresql"] = userConfigTestSchema().Properties["rsyslog_user_config"]

	t.Run("expand sends the legacy block", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithTestPlan(map[string]any{}))
		require.NoError(t, err)
		dto := map[string]any{
			"endpoint_type":       "external_postgresql",
			"external_postgresql": map[string]any{"server": "pg.example.com"},
			"datadog_user_config": map[string]any{"site": "datadoghq.com"},
		}
		require.NoError(t, ExpandUserConfig("endpoint_type")(d, dto))
		assert.Equal(t, map[string]any{
			"endpoint_type": "external_postgresql",
			"user_config":   map[string]any{"server": "pg.example.com"},
		}, dto)
	})

	t.Run("flatten sets the legacy block", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithIsDataSource())
		require.NoError(t, err)
		dto := map[string]any{"endpoint_type": "external_postgresql", "user_config": map[string]any{"server": "pg.example.com"}}
		require.NoError(t, FlattenUserConfig("endpoint_type")(d, dto))
		assert.Equal(t, map[string]any{"endpoint_type": "external_postgresql", "external_postgresql": map[string]any{"server": "pg.example.com"}}, dto)
	})

	t.Run("flatten skips a type without a block", func(t *testing.T) {
		t.Parallel()
		d, err := NewResourceData(sch, nil, WithIsDataSource())
		require.NoError(t, err)
		dto := map[string]any{"endpoint_type": "jolokia", "user_config": map[string]any{"server": "x"}}
		require.NoError(t, FlattenUserConfig("endpoint_type")(d, dto))
		assert.Equal(t, map[string]any{"endpoint_type": "jolokia"}, dto)
	})
}

func TestKeepSetValues(t *testing.T) {
	t.Parallel()

	sch := userConfigTestSchema().Properties["datadog_user_config"]
	rsp := map[string]any{"site": "datadoghq.com", "opensearch": map[string]any{"index_stats_enabled": true}}

	t.Run("nested block not set", func(t *testing.T) {
		t.Parallel()
		got, ok := keepSetValues(sch, rsp, []any{map[string]any{"site": "x"}})
		require.True(t, ok)
		assert.Equal(t, map[string]any{"site": "datadoghq.com"}, got)
	})

	t.Run("no known value", func(t *testing.T) {
		t.Parallel()
		_, ok := keepSetValues(sch, rsp, nil)
		assert.False(t, ok)
	})

	t.Run("nil response value", func(t *testing.T) {
		t.Parallel()
		got, ok := keepSetValues(sch, map[string]any{"site": nil}, []any{map[string]any{"site": "x"}})
		require.True(t, ok)
		assert.Equal(t, map[string]any{}, got)
	})

	t.Run("no schema", func(t *testing.T) {
		t.Parallel()
		_, ok := keepSetValues(nil, rsp, rsp)
		assert.False(t, ok)
	})

	t.Run("list objects without a known one are dropped", func(t *testing.T) {
		t.Parallel()
		tables := sch.Items.Properties["tables"]
		got, ok := keepSetValues(tables, []any{
			map[string]any{"name": "foo", "auto_offset_reset": "earliest"},
			map[string]any{"name": "bar", "auto_offset_reset": "earliest"},
		}, []any{map[string]any{"name": "foo"}})
		require.True(t, ok)
		assert.Equal(t, []any{map[string]any{"name": "foo"}}, got)
	})
}
