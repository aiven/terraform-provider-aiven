package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestKafkaConnect(t *testing.T) {
	t.Parallel()
	runCompat(t, kafkaConnectFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"kafka_connect_user_config": nil},
			wantPlan: []string{"kafka_connect_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"kafka_connect_user_config": []any{map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
			}}},
			wantPlan: []string{
				"kafka_connect_user_config[0].kafka_connect",
				"kafka_connect_user_config[0].public_access",
				"state",
			},
		},
	})
}

func kafkaConnectFixture() fixture {
	const (
		host = "my-kafkaconnect.example.com"
		uri  = "https://avnadmin:secret@my-kafkaconnect.example.com:443"
	)

	return fixture{
		typeName: "aiven_kafka_connect",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "kafka_connect", "connection_uri": host + ":443", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 443, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "kafka_connect", "connection_uri": host + ":443", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 443, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "replica",
				},
			},
			"disk_space":                 nil,
			"disk_space_cap":             "0B",
			"disk_space_default":         "0B",
			"disk_space_step":            "0B",
			"disk_space_used":            "0B",
			"id":                         testProject + "/my-kafkaconnect",
			"kafka_connect_user_config":  []any{kafkaConnectUserConfigState(nil)},
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-kafkaconnect",
			"service_password":           "secret",
			"service_port":               443,
			"service_type":               "kafka_connect",
			"service_uri":                uri,
			"service_username":           "avnadmin",
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"timeouts":                   nil,
		},
		config: map[string]any{
			"cloud_name":                testCloud,
			"kafka_connect_user_config": []any{kafkaConnectUserConfig(nil)},
			"plan":                      testPlan,
			"project":                   testProject,
			"service_name":              "my-kafkaconnect",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{Component: "kafka_connect", Host: host, Port: 443, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
				{Component: "kafka_connect", Host: host, Port: 443, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypeReplica},
			},
			ConnectionInfo: &service.ConnectionInfoOut{},
			Maintenance:    &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:           testPlan,
			ServiceName:    "my-kafkaconnect",
			ServiceType:    "kafka_connect",
			ServiceUri:     uri,
			ServiceUriParams: map[string]string{
				"host": host, "port": "443", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
				"kafka_connect": map[string]any{
					"consumer_isolation_level": "read_committed",
				},
				"public_access": map[string]any{
					"kafka_connect": true,
					"prometheus":    false,
				},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 0, DiskSpaceCapMb: new(0), DiskSpaceStepMb: new(0),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "kafka_connect",
		},
	}
}

func kafkaConnectUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
		"kafka_connect": []any{map[string]any{
			"consumer_isolation_level": "read_committed",
		}},
		"public_access": []any{map[string]any{
			"kafka_connect": true,
			"prometheus":    false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func kafkaConnectUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions": nil,
		"custom_plugins":            []any{},
		"gcp_auth_allowed_urls":     nil,
		"ip_filter":                 []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":          []any{},
		"ip_filter_string":          nil,
		"kafka_connect": []any{map[string]any{
			"connector_client_config_override_policy": "",
			"consumer_auto_offset_reset":              "",
			"consumer_fetch_max_bytes":                0,
			"consumer_isolation_level":                "read_committed",
			"consumer_max_partition_fetch_bytes":      0,
			"consumer_max_poll_interval_ms":           0,
			"consumer_max_poll_records":               0,
			"offset_flush_interval_ms":                0,
			"offset_flush_timeout_ms":                 0,
			"prefer_ipv6_address_enable":              false,
			"producer_batch_size":                     0,
			"producer_buffer_memory":                  0,
			"producer_compression_type":               "",
			"producer_linger_ms":                      0,
			"producer_max_request_size":               0,
			"scheduled_rebalance_max_delay_ms":        0,
			"session_timeout_ms":                      0,
		}},
		"plugin_versions":    []any{},
		"preferred_zones":    nil,
		"private_access":     []any{},
		"privatelink_access": []any{},
		"public_access": []any{map[string]any{
			"kafka_connect": true,
			"prometheus":    false,
		}},
		"sasl_oauthbearer_allowed_urls": nil,
		"secret_providers":              []any{},
		"service_log":                   false,
		"static_ips":                    false,
	}
	maps.Copy(result, options)
	return result
}
