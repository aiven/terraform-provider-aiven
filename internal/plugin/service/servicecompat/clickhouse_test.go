package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestClickHouse(t *testing.T) {
	t.Parallel()
	runCompat(t, clickhouseFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"clickhouse_user_config": nil},
			wantPlan: []string{"clickhouse_user_config", "state"},
		},
		{
			name: "config has no ip_filter",
			config: map[string]any{"clickhouse_user_config": []any{map[string]any{
				"clickhouse_version": "25.8",
				"backup_hour":        3,
				"backup_minute":      30,
			}}},
		},
	})
}

func clickhouseFixture() fixture {
	const (
		host = "my-clickhouse.example.com"
		uri  = "clickhouse://avnadmin:secret@my-clickhouse.example.com:12345/default"
	)

	return fixture{
		typeName: "aiven_clickhouse",
		version:  0,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"clickhouse": []any{map[string]any{
				"uris": []any{uri},
			}},
			"clickhouse_user_config": []any{clickhouseUserConfigState(nil)},
			"components": []any{map[string]any{
				"component": "clickhouse", "connection_uri": host + ":12345", "host": host,
				"kafka_authentication_method": "", "kafka_ssl_ca": "",
				"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
			}},
			"disk_space":                 nil,
			"disk_space_cap":             "400GiB",
			"disk_space_default":         "80GiB",
			"disk_space_step":            "10GiB",
			"disk_space_used":            "80GiB",
			"id":                         testProject + "/my-clickhouse",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-clickhouse",
			"service_password":           "secret",
			"service_port":               12345,
			"service_type":               "clickhouse",
			"service_uri":                host + ":12345",
			"service_username":           "avnadmin",
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"timeouts":                   nil,
		},
		config: map[string]any{
			"cloud_name":             testCloud,
			"clickhouse_user_config": []any{clickhouseUserConfig(nil)},
			"plan":                   testPlan,
			"project":                testProject,
			"service_name":           "my-clickhouse",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{{
				Component: "clickhouse", Host: host, Port: 12345,
				Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary,
			}},
			ConnectionInfo: &service.ConnectionInfoOut{Clickhouse: []string{uri}},
			DiskSpaceMb:    new(81920),
			Maintenance:    &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:           testPlan,
			ServiceName:    "my-clickhouse",
			ServiceType:    "clickhouse",
			ServiceUri:     host + ":12345",
			ServiceUriParams: map[string]string{
				"host": host, "port": "12345", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"backup_hour": 3, "backup_minute": 30,
				"clickhouse_version": "25.8",
				"ip_filter":          []any{"0.0.0.0/0", "::/0"},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 81920, DiskSpaceCapMb: new(409600), DiskSpaceStepMb: new(10240),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "clickhouse",
		},
	}
}

func clickhouseUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"backup_hour": 3, "backup_minute": 30,
		"clickhouse_version": "25.8",
		"ip_filter":          []any{"0.0.0.0/0", "::/0"},
	}
	maps.Copy(result, options)
	return result
}

func clickhouseUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"backup_hour":                3,
		"backup_minute":              30,
		"clickhouse_version":         "25.8",
		"enable_ipv6":                false,
		"ip_filter":                  []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":           []any{},
		"ip_filter_string":           nil,
		"private_access":             []any{},
		"privatelink_access":         []any{},
		"project_to_fork_from":       "",
		"public_access":              []any{},
		"recovery_basebackup_name":   "",
		"server_settings":            []any{},
		"service_log":                false,
		"service_to_fork_from":       "",
		"session_settings":           []any{},
		"static_ips":                 false,
		"tiered_storage_move_factor": 0.0,
	}
	maps.Copy(result, options)
	return result
}
