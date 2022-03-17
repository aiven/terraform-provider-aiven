package schemautil

import (
	"reflect"
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

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
					DiffSuppressFunc: CreateOnlyDiffSuppressFunc,
					Description:      "Custom password for admin user",
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
			false,
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

func Test_convertTerraformUserConfigToAPICompatibleFormat(t *testing.T) {
	entrySchema := templates.GetUserConfigSchema("common")["kafka"].(map[string]interface{})
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
						"auto_create_topics_enable":       "true",
						"connections_max_idle_ms":         1001,
						"group_max_session_timeout_ms":    300000,
						"group_min_session_timeout_ms":    6000,
						"log_cleaner_min_cleanable_ratio": 0.5,
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
