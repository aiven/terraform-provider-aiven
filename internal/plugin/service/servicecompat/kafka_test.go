package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestKafka(t *testing.T) {
	t.Parallel()
	runCompat(t, kafkaFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"kafka_user_config": nil},
			wantPlan: []string{"kafka_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"kafka_user_config": []any{map[string]any{
				"kafka_version":   "3.9",
				"schema_registry": true,
			}}},
			wantPlan: []string{
				"kafka_user_config[0].kafka_authentication_methods",
				"state",
			},
		},
		{
			// The default, which the API doesn't have
			name:   "default_acl unset",
			state:  map[string]any{"default_acl": true},
			config: map[string]any{"default_acl": nil},
		},
	})
}

func kafkaFixture() fixture {
	const (
		host     = "my-kafka.example.com"
		port     = 12345
		portRest = 12348
		uri      = "kafka+ssl://my-kafka.example.com:12345"
		restURI  = "https://my-kafka.example.com:12348"
		srURI    = "https://my-kafka.example.com:12348"
	)

	return fixture{
		typeName: "aiven_kafka",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "kafka", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "certificate", "kafka_ssl_ca": "project_ca",
					"port": port, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "kafka_rest", "connection_uri": host + ":12348", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": portRest, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
			},
			"default_acl":        false,
			"disk_space":         nil,
			"disk_space_cap":     "0B",
			"disk_space_default": "90GiB",
			"disk_space_step":    "0B",
			"disk_space_used":    "90GiB",
			"id":                 testProject + "/my-kafka",
			"kafka": []any{map[string]any{
				"access_cert": "CERT", "access_key": "KEY", "connect_uri": "",
				"rest_uri": restURI, "schema_registry_uri": srURI,
				"uris": []any{uri, uri},
			}},
			"kafka_user_config":          []any{kafkaUserConfigState(nil)},
			"karapace":                   nil,
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-kafka",
			"service_password":           nil,
			"service_port":               port,
			"service_type":               "kafka",
			"service_uri":                uri,
			"service_username":           nil,
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"timeouts":                   nil,
		},
		config: map[string]any{
			"cloud_name":        testCloud,
			"default_acl":       false,
			"kafka_user_config": []any{kafkaUserConfig(nil)},
			"plan":              testPlan,
			"project":           testProject,
			"service_name":      "my-kafka",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{
					Component:                 "kafka",
					Host:                      host,
					Port:                      port,
					KafkaAuthenticationMethod: "certificate",
					KafkaSslCa:                "project_ca",
					Route:                     service.RouteTypeDynamic,
					Ssl:                       new(true),
					Usage:                     service.UsageTypePrimary,
				},
				{
					Component: "kafka_rest",
					Host:      host,
					Port:      portRest,
					Route:     service.RouteTypeDynamic,
					Ssl:       new(true),
					Usage:     service.UsageTypePrimary,
				},
			},
			ConnectionInfo: &service.ConnectionInfoOut{
				Kafka:             []string{uri, uri},
				KafkaAccessCert:   new("CERT"),
				KafkaAccessKey:    new("KEY"),
				KafkaRestUri:      new(restURI),
				SchemaRegistryUri: new(srURI),
			},
			DiskSpaceMb: new(92160),
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-kafka",
			ServiceType: "kafka",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"host": host,
				"port": "12345",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter":       []any{"0.0.0.0/0", "::/0"},
				"kafka_version":   "3.9",
				"schema_registry": true,
				"kafka_authentication_methods": map[string]any{
					"certificate": true,
					"sasl":        false,
				},
			},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb:     92160,
			DiskSpaceCapMb:  new(0),
			DiskSpaceStepMb: new(0),
			NodeCount:       1,
			ServicePlan:     testPlan,
			ServiceType:     "kafka",
		},
	}
}

func kafkaUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter":       []any{"0.0.0.0/0", "::/0"},
		"kafka_version":   "3.9",
		"schema_registry": true,
		"kafka_authentication_methods": []any{map[string]any{
			"certificate": true,
			"sasl":        false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func kafkaUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":  nil,
		"aiven_kafka_topic_messages": false,
		"backup_interval_hours":      0,
		"backup_retention_days":      0,
		"custom_domain":              "",
		"enable_ipv6":                false,
		"follower_fetching":          []any{},
		"gcp_auth_allowed_urls":      nil,
		"inkless":                    []any{},
		"ip_filter":                  []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":           []any{},
		"ip_filter_string":           nil,
		"kafka":                      []any{},
		"kafka_authentication_methods": []any{map[string]any{
			"certificate": true,
			"sasl":        false,
		}},
		"kafka_connect":                  false,
		"kafka_connect_config":           []any{},
		"kafka_connect_plugin_versions":  []any{},
		"kafka_connect_secret_providers": []any{},
		"kafka_diskless":                 []any{},
		"kafka_rest":                     false,
		"kafka_rest_authorization":       false,
		"kafka_rest_config":              []any{},
		"kafka_sasl_mechanisms":          []any{},
		"kafka_version":                  "3.9",
		"karapace_version":               "",
		"letsencrypt_sasl":               false,
		"letsencrypt_sasl_privatelink":   false,
		"preferred_zones":                nil,
		"private_access":                 []any{},
		"privatelink_access":             []any{},
		"public_access":                  []any{},
		"sasl_oauthbearer_allowed_urls":  nil,
		"schema_registry":                true,
		"schema_registry_config":         []any{},
		"service_log":                    false,
		"single_zone":                    []any{},
		"static_ips":                     false,
		"tiered_storage":                 []any{},
	}
	maps.Copy(result, options)
	return result
}
