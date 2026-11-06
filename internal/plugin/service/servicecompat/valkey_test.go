package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestValkey(t *testing.T) {
	t.Parallel()
	runCompat(t, valkeyFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"valkey_user_config": nil},
			wantPlan: []string{"valkey_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"valkey_user_config": []any{map[string]any{
				"valkey_version": "7.2",
			}}},
			wantPlan: []string{
				"valkey_user_config[0].public_access",
				"state",
			},
		},
		{
			name: "tech_emails in state",
			state: map[string]any{
				"tech_emails": []any{map[string]any{"email": "me@example.com"}},
			},
			config: map[string]any{
				"tech_emails": []any{map[string]any{"email": "me@example.com"}},
			},
			api: func(s *service.ServiceGetOut) {
				s.TechEmails = []service.TechEmailOut{{Email: "me@example.com"}}
			},
		},
	})
}

func valkeyFixture() fixture {
	const (
		host = "my-valkey.example.com"
		uri  = "rediss://avnadmin:secret@my-valkey.example.com:12345"
	)

	return fixture{
		typeName: "aiven_valkey",
		version:  0,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "valkey", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "valkey", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "replica",
				},
			},
			"disk_space":                 nil,
			"disk_space_cap":             "0B",
			"disk_space_default":         "0B",
			"disk_space_step":            "0B",
			"disk_space_used":            "0B",
			"id":                         testProject + "/my-valkey",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-valkey",
			"service_password":           "secret",
			"service_port":               12345,
			"service_type":               "valkey",
			"service_uri":                uri,
			"service_username":           "avnadmin",
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"timeouts":                   nil,
			"valkey": []any{map[string]any{
				"password":    "secret",
				"replica_uri": "",
				"slave_uris":  []any{},
				"uris":        []any{uri},
			}},
			"valkey_user_config": []any{valkeyUserConfigState(nil)},
		},
		config: map[string]any{
			"cloud_name":         testCloud,
			"valkey_user_config": []any{valkeyUserConfig(nil)},
			"plan":               testPlan,
			"project":            testProject,
			"service_name":       "my-valkey",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{Component: "valkey", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
				{Component: "valkey", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypeReplica},
			},
			ConnectionInfo: &service.ConnectionInfoOut{
				Valkey:         []string{uri},
				ValkeyPassword: new("secret"),
			},
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-valkey",
			ServiceType: "valkey",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"host": host, "port": "12345", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"frequent_snapshots":                   true,
				"ip_filter":                            []any{"0.0.0.0/0", "::/0"},
				"valkey_active_defrag_ignore_bytes":    104857600,
				"valkey_active_defrag_threshold_lower": 10,
				"valkey_active_expire_effort":          1,
				"valkey_io_threads":                    1,
				"valkey_lfu_decay_time":                1,
				"valkey_lfu_log_factor":                10,
				"valkey_maxmemory_policy":              "noeviction",
				"valkey_ssl":                           true,
				"valkey_timeout":                       300,
				"valkey_version":                       "7.2",
				"public_access":                        map[string]any{"valkey": true, "prometheus": false},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 0, DiskSpaceCapMb: new(0), DiskSpaceStepMb: new(0),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "valkey",
		},
	}
}

func valkeyUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"frequent_snapshots":      true,
		"ip_filter":               []any{"0.0.0.0/0", "::/0"},
		"valkey_maxmemory_policy": "noeviction",
		"valkey_ssl":              true,
		"valkey_timeout":          300,
		"valkey_version":          "7.2",
		"public_access":           []any{map[string]any{"valkey": true, "prometheus": false}},
	}
	maps.Copy(result, options)
	return result
}

func valkeyUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":                nil,
		"backup_hour":                              0,
		"backup_minute":                            0,
		"enable_ipv6":                              false,
		"frequent_snapshots":                       true,
		"ip_filter":                                []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":                         []any{},
		"ip_filter_string":                         nil,
		"migration":                                []any{},
		"private_access":                           []any{},
		"privatelink_access":                       []any{},
		"project_to_fork_from":                     "",
		"public_access":                            []any{map[string]any{"valkey": true, "prometheus": false}},
		"recovery_basebackup_name":                 "",
		"replicas":                                 0,
		"service_log":                              false,
		"service_to_fork_from":                     "",
		"shard_count":                              0,
		"static_ips":                               false,
		"valkey_acl_channels_default":              "",
		"valkey_active_defrag_ignore_bytes":        104857600,
		"valkey_active_defrag_threshold_lower":     10,
		"valkey_active_expire_effort":              1,
		"valkey_activedefrag":                      false,
		"valkey_io_threads":                        1,
		"valkey_lfu_decay_time":                    1,
		"valkey_lfu_log_factor":                    10,
		"valkey_maxmemory_policy":                  "noeviction",
		"valkey_notify_keyspace_events":            "",
		"valkey_number_of_databases":               0,
		"valkey_persistence":                       "",
		"valkey_pubsub_client_output_buffer_limit": 0,
		"valkey_ssl":                               true,
		"valkey_timeout":                           300,
		"valkey_version":                           "7.2",
	}
	maps.Copy(result, options)
	return result
}
