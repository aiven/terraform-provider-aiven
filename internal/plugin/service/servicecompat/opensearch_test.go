package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestOpenSearch(t *testing.T) {
	t.Parallel()
	runCompat(t, opensearchFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"opensearch_user_config": nil},
			wantPlan: []string{"opensearch_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"opensearch_user_config": []any{map[string]any{
				"opensearch_version": "2.17",
			}}},
			wantPlan: []string{
				"opensearch_user_config[0].index_patterns",
				"opensearch_user_config[0].opensearch_dashboards",
				"opensearch_user_config[0].public_access",
				"state",
			},
		},
	})
}

func opensearchFixture() fixture {
	const (
		host    = "my-opensearch.example.com"
		osURI   = "https://avnadmin:secret@my-opensearch.example.com:12345"
		dashURI = "https://my-opensearch.example.com:12345/_dashboards"
	)

	return fixture{
		typeName: "aiven_opensearch",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "opensearch", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "opensearch", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "replica",
				},
			},
			"disk_space":                 nil,
			"disk_space_cap":             "400GiB",
			"disk_space_default":         "80GiB",
			"disk_space_step":            "10GiB",
			"disk_space_used":            "80GiB",
			"id":                         testProject + "/my-opensearch",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"opensearch": []any{map[string]any{
				"kibana_uri":                "",
				"opensearch_dashboards_uri": dashURI,
				"password":                  "secret",
				"uris":                      []any{osURI},
				"username":                  "avnadmin",
			}},
			"opensearch_user_config": []any{opensearchUserConfigState(nil)},
			"plan":                   testPlan,
			"project":                testProject,
			"project_vpc_id":         "",
			"service_host":           host,
			"service_integrations":   []any{},
			"service_name":           "my-opensearch",
			"service_password":       "secret",
			"service_port":           12345,
			"service_type":           "opensearch",
			"service_uri":            osURI,
			"service_username":       "avnadmin",
			"state":                  "RUNNING",
			"static_ips":             nil,
			"tag":                    []any{},
			"tech_emails":            []any{},
			"termination_protection": false,
			"timeouts":               nil,
		},
		config: map[string]any{
			"cloud_name":             testCloud,
			"opensearch_user_config": []any{opensearchUserConfig(nil)},
			"plan":                   testPlan,
			"project":                testProject,
			"service_name":           "my-opensearch",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{Component: "opensearch", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
				{Component: "opensearch", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypeReplica},
			},
			ConnectionInfo: &service.ConnectionInfoOut{
				Opensearch:              []string{osURI},
				OpensearchDashboardsUri: new(dashURI),
				OpensearchUsername:      new("avnadmin"),
				OpensearchPassword:      new("secret"),
			},
			DiskSpaceMb: new(81920),
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-opensearch",
			ServiceType: "opensearch",
			ServiceUri:  osURI,
			ServiceUriParams: map[string]string{
				"host": host, "port": "12345", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter":          []any{"0.0.0.0/0", "::/0"},
				"opensearch_version": "2.17",
				"index_patterns": []any{
					map[string]any{"pattern": "logs-*", "max_index_count": 3, "sorting_algorithm": "alphabetical"},
				},
				"opensearch_dashboards": map[string]any{
					"enabled":                      true,
					"max_old_space_size":           128,
					"multiple_data_source_enabled": true,
					"opensearch_request_timeout":   30000,
					"session_keepalive":            true,
					"session_ttl":                  "24h",
				},
				"public_access": map[string]any{
					"opensearch": true, "opensearch_dashboards": true, "prometheus": false,
				},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 81920, DiskSpaceCapMb: new(409600), DiskSpaceStepMb: new(10240),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "opensearch",
		},
	}
}

func opensearchUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter":          []any{"0.0.0.0/0", "::/0"},
		"opensearch_version": "2.17",
		"index_patterns": []any{map[string]any{
			"pattern": "logs-*", "max_index_count": 3, "sorting_algorithm": "alphabetical",
		}},
		"opensearch_dashboards": []any{map[string]any{
			"enabled": true, "session_ttl": "24h",
		}},
		"public_access": []any{map[string]any{
			"opensearch": true, "opensearch_dashboards": true, "prometheus": false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func opensearchUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":             nil,
		"azure_migration":                       []any{},
		"custom_domain":                         "",
		"disable_replication_factor_adjustment": false,
		"gcs_migration":                         []any{},
		"index_patterns": []any{
			map[string]any{"pattern": "logs-*", "max_index_count": 3, "sorting_algorithm": "alphabetical"},
		},
		"index_rollup":                []any{},
		"index_template":              []any{},
		"ip_filter":                   []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":            []any{},
		"ip_filter_string":            nil,
		"jwt":                         []any{},
		"keep_index_refresh_interval": false,
		"max_index_count":             0,
		"openid":                      []any{},
		"opensearch":                  []any{},
		"opensearch_dashboards": []any{map[string]any{
			"enabled": true, "max_old_space_size": 128, "multiple_data_source_enabled": true,
			"opensearch_request_timeout": 30000, "session_keepalive": true, "session_ttl": "24h",
		}},
		"opensearch_version":   "2.17",
		"private_access":       []any{},
		"privatelink_access":   []any{},
		"project_to_fork_from": "",
		"public_access": []any{map[string]any{
			"opensearch": true, "opensearch_dashboards": true, "prometheus": false,
		}},
		"recovery_basebackup_name": "",
		"s3_migration":             []any{},
		"saml":                     []any{},
		"service_log":              false,
		"service_to_fork_from":     "",
		"static_ips":               false,
	}
	maps.Copy(result, options)
	return result
}
