package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_readUserConfigJSONSchema(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		args  args
		panic bool
	}{
		{
			"basic",
			args{name: "service_user_config_schema.json"},
			false,
		},
		{
			"wrong-file-name",
			args{name: "wrong-file-name.json"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Panics(t,
					func() { readUserConfigJSONSchema(tt.args.name) },
					"should panic")
			} else {
				if got := readUserConfigJSONSchema(tt.args.name); !assert.NotEmpty(t, got) {
					t.Errorf("readUserConfigJSONSchema() = %v is empty", got)
				}
			}
		})
	}
}

func TestGetUserConfigSchema(t *testing.T) {
	type args struct {
		resourceType string
	}
	tests := []struct {
		name  string
		args  args
		panic bool
	}{
		{
			"endpoint",
			args{resourceType: "endpoint"},
			false,
		},
		{
			"integration",
			args{resourceType: "integration"},
			false,
		},
		{
			"service",
			args{resourceType: "service"},
			false,
		},
		{
			"wrong-type",
			args{resourceType: "wrong"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Panics(t,
					func() { GetUserConfigSchema(tt.args.resourceType) },
					"should panic")
			} else {
				if got := GetUserConfigSchema(tt.args.resourceType); !assert.NotEmpty(t, got) {
					t.Errorf("GetUserConfigSchema() = %v is empty", got)
				}
			}
		})
	}
}

func TestGenerateTerraformUserConfigSchema(t *testing.T) {
	type args struct {
		data map[string]interface{}
	}
	var tests = []struct {
		name  string
		args  args
		want  map[string]*schema.Schema
		panic bool
	}{
		{
			"basic",
			args{
				data: map[string]interface{}{
					"properties": map[string]interface{}{
						"admin_password": map[string]interface{}{
							"createOnly": true,
							"default":    nil,
							"example":    "z66o9QXqKM",
							"maxLength":  256,
							"minLength":  8,
							"testFloat":  9.9,
							"pattern":    "^[a-zA-Z0-9-_]+$",
							"title":      "Custom password for admin user",
							"type": []interface{}{
								"string",
								"null",
							},
							"user_error": "Must consist of alpha-numeric characters, underscores or dashes",
						},
					},
				},
			},
			map[string]*schema.Schema{
				"admin_password": {
					Type:             schema.TypeString,
					ConfigMode:       0,
					Optional:         true,
					Required:         false,
					Computed:         false,
					Sensitive:        true,
					DiffSuppressFunc: createOnlyDiffSuppressFunc,
					Description:      "Custom password for admin user",
					Default:          "<<value not set>>",
				},
			},
			false,
		},
		{
			"no-type",
			args{
				data: map[string]interface{}{
					"properties": map[string]interface{}{
						"admin_password": map[string]interface{}{
							"createOnly": true,
							"default":    nil,
							"example":    "z66o9QXqKM",
							"maxLength":  256,
							"minLength":  8,
							"pattern":    "^[a-zA-Z0-9-_]+$",
							"title":      "Custom password for admin user",
							"user_error": "Must consist of alpha-numeric characters, underscores or dashes",
						},
					},
				},
			},
			nil,
			true,
		},
		{
			// type should be a string or []interface{}
			"wrong-type",
			args{
				data: map[string]interface{}{
					"properties": map[string]interface{}{
						"admin_password": map[string]interface{}{
							"createOnly": true,
							"default":    nil,
							"example":    "z66o9QXqKM",
							"maxLength":  256,
							"minLength":  8,
							"pattern":    "^[a-zA-Z0-9-_]+$",
							"title":      "Custom password for admin user",
							"type":       123,
							"user_error": "Must consist of alpha-numeric characters, underscores or dashes",
						},
					},
				},
			},
			nil,
			true,
		},
		{
			"no-properties",
			args{
				data: map[string]interface{}{
					"admin_password": map[string]interface{}{
						"createOnly": true,
						"default":    nil,
						"example":    "z66o9QXqKM",
						"maxLength":  256,
						"minLength":  8,
						"pattern":    "^[a-zA-Z0-9-_]+$",
						"title":      "Custom password for admin user",
						"type": []interface{}{
							"string",
							"null",
						},
						"user_error": "Must consist of alpha-numeric characters, underscores or dashes",
					},
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wantKeys []string
			var gotKeys []string
			var got map[string]*schema.Schema

			if tt.panic {
				if !assert.Panics(t,
					func() {
						got = GenerateTerraformUserConfigSchema(tt.args.data)
					},
					"should panic") {
					t.FailNow()
				}
			} else {
				got = GenerateTerraformUserConfigSchema(tt.args.data)
			}

			for k := range tt.want {
				wantKeys = append(wantKeys, k)
			}

			for k := range got {
				gotKeys = append(gotKeys, k)
			}

			if !reflect.DeepEqual(gotKeys, wantKeys) {
				t.Errorf("GenerateTerraformUserConfigSchema() keys validation error = %+#v, want %+#v", gotKeys, wantKeys)
			}

			for k, shema := range got {
				assert.NotEmpty(t, k)
				assert.Equal(t, shema.GoString(), tt.want[k].GoString())
			}
		})
	}
}

func TestConvertAPIUserConfigToTerraformCompatibleFormat(t *testing.T) {
	type args struct {
		configType string
		entryType  string
		userConfig map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  []map[string]interface{}
		panic bool
	}{
		{
			"basic-kafka",
			args{
				configType: "service",
				entryType:  "kafka",
				userConfig: map[string]interface{}{
					"ip_filter": []interface{}{
						"0.0.0.0/0",
					},
					"kafka": map[string]interface{}{
						"auto_create_topics_enable":    true,
						"connections_max_idle_ms":      int32(10),
						"group_max_session_timeout_ms": int32(300000),
					},
					"kafka_authentication_methods": map[string]interface{}{
						"certificate": true,
						"sasl":        false,
					},
					"kafka_connect":   false,
					"kafka_rest":      false,
					"schema_registry": false,
					"kafka_version":   "2.1",
				},
			},
			[]map[string]interface{}{
				{
					"ip_filter": []interface{}{
						"0.0.0.0/0",
					},
					"kafka": []map[string]interface{}{
						{
							"log_cleanup_policy":                         "delete",
							"auto_create_topics_enable":                  "true",
							"compression_type":                           "<<value not set>>",
							"connections_max_idle_ms":                    10,
							"default_replication_factor":                 -1,
							"group_max_session_timeout_ms":               300000,
							"group_min_session_timeout_ms":               float64(6000),
							"log_cleaner_max_compaction_lag_ms":          -1,
							"log_cleaner_min_cleanable_ratio":            0.5,
							"log_cleaner_min_compaction_lag_ms":          -1,
							"log_message_timestamp_difference_max_ms":    -1,
							"log_message_timestamp_type":                 "<<value not set>>",
							"log_retention_bytes":                        -1,
							"log_retention_hours":                        -1,
							"log_segment_bytes":                          -1,
							"max_connections_per_ip":                     -1,
							"message_max_bytes":                          1.000012e+06,
							"num_partitions":                             -1,
							"offsets_retention_minutes":                  float64(1440),
							"producer_purgatory_purge_interval_requests": -1,
							"replica_fetch_max_bytes":                    -1,
							"replica_fetch_response_max_bytes":           -1,
							"socket_request_max_bytes":                   -1,
						},
					},
					"kafka_authentication_methods": []map[string]interface{}{
						{
							"certificate": true,
							"sasl":        false,
						},
					},
					"kafka_connect":        false,
					"kafka_connect_config": []map[string]interface{}{},
					"kafka_rest":           false,
					"kafka_rest_config":    []map[string]interface{}{},
					"kafka_version":        "2.1",
					"private_access":       []map[string]interface{}{},
					"public_access":        []map[string]interface{}{},
					"schema_registry":      false,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertAPIUserConfigToTerraformCompatibleFormat(tt.args.configType, tt.args.entryType, tt.args.userConfig)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_convertTerraformUserConfigToAPICompatibleFormat1(t *testing.T) {
	entrySchema := GetUserConfigSchema("service")["kafka"].(map[string]interface{})
	entrySchemaProps := entrySchema["properties"].(map[string]interface{})

	type args struct {
		serviceType  string
		newResource  bool
		userConfig   map[string]interface{}
		configSchema map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			"basic",
			args{
				serviceType: "kafka",
				newResource: false,
				userConfig: map[string]interface{}{
					"ip_filter": []interface{}{
						"0.0.0.0/0",
					},
					"kafka": map[string]interface{}{
						"auto_create_topics_enable":                  "true",
						"compression_type":                           "<<value not set>>",
						"connections_max_idle_ms":                    1001,
						"default_replication_factor":                 -1,
						"group_max_session_timeout_ms":               300000,
						"group_min_session_timeout_ms":               6000,
						"log_cleaner_max_compaction_lag_ms":          -1,
						"log_cleaner_min_cleanable_ratio":            0.5,
						"log_cleaner_min_compaction_lag_ms":          -1,
						"log_message_timestamp_difference_max_ms":    -1,
						"log_message_timestamp_type":                 "<<value not set>>",
						"log_retention_bytes":                        -1,
						"log_retention_hours":                        -1,
						"log_segment_bytes":                          -1,
						"max_connections_per_ip":                     -1,
						"message_max_bytes":                          1,
						"num_partitions":                             -1,
						"offsets_retention_minutes":                  1440,
						"producer_purgatory_purge_interval_requests": -1,
						"replica_fetch_max_bytes":                    -1,
						"replica_fetch_response_max_bytes":           -1,
						"socket_request_max_bytes":                   -1,
					},
					"kafka_authentication_methods": map[string]interface{}{
						"certificate": true,
						"sasl":        false,
					},
					"kafka_connect":        false,
					"kafka_connect_config": map[string]interface{}{},
					"kafka_rest":           false,
					"kafka_rest_config":    map[string]interface{}{},
					"kafka_version":        "2.1",
					"private_access":       map[string]interface{}{},
					"public_access":        map[string]interface{}{},
					"schema_registry":      false,
				},
				configSchema: entrySchemaProps,
			},
			map[string]interface{}{
				"ip_filter": []interface{}{
					"0.0.0.0/0",
				},
				"kafka": map[string]interface{}{
					"auto_create_topics_enable":       true,
					"connections_max_idle_ms":         1001,
					"group_max_session_timeout_ms":    300000,
					"group_min_session_timeout_ms":    6000,
					"log_cleaner_min_cleanable_ratio": 0.5,
					"log_retention_bytes":             -1,
					"log_retention_hours":             -1,
					"message_max_bytes":               1,
					"offsets_retention_minutes":       1440,
				},
				"kafka_authentication_methods": map[string]interface{}{
					"certificate": true,
					"sasl":        false,
				},
				"kafka_connect":        false,
				"kafka_connect_config": map[string]interface{}{},
				"kafka_rest":           false,
				"kafka_rest_config":    map[string]interface{}{},
				"kafka_version":        "2.1",
				"private_access":       map[string]interface{}{},
				"public_access":        map[string]interface{}{},
				"schema_registry":      false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTerraformUserConfigToAPICompatibleFormat(tt.args.serviceType, tt.args.newResource, tt.args.userConfig, tt.args.configSchema)
			assert.Equal(t, got, tt.want)
		})
	}
}
