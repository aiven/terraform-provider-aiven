package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestMySQL(t *testing.T) {
	t.Parallel()
	runCompat(t, mysqlFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"mysql_user_config": nil},
			wantPlan: []string{"mysql_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"mysql_user_config": []any{map[string]any{
				"mysql_version": "8.0.35",
				"backup_hour":   5,
				"backup_minute": 19,
			}}},
			wantPlan: []string{
				"mysql_user_config[0].mysql",
				"mysql_user_config[0].public_access",
				"state",
			},
		},
	})
}

func mysqlFixture() fixture {
	const (
		host = "my-mysql.example.com"
		uri  = "mysql://avnadmin:secret@my-mysql.example.com:12345/defaultdb?ssl-mode=REQUIRED"
	)

	return fixture{
		typeName: "aiven_mysql",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "mysql", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "mysql", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "replica",
				},
			},
			"disk_space":                 nil,
			"disk_space_cap":             "400GiB",
			"disk_space_default":         "80GiB",
			"disk_space_step":            "10GiB",
			"disk_space_used":            "80GiB",
			"id":                         testProject + "/my-mysql",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"mysql": []any{map[string]any{
				"params": []any{map[string]any{
					"database_name": "defaultdb", "host": host, "password": "secret",
					"port": 12345, "sslmode": "REQUIRED", "user": "avnadmin",
				}},
				"replica_uri":  "",
				"standby_uris": []any{},
				"syncing_uris": []any{},
				"uris":         []any{uri},
			}},
			"mysql_user_config":      []any{mysqlUserConfigState(nil)},
			"plan":                   testPlan,
			"project":                testProject,
			"project_vpc_id":         "",
			"service_host":           host,
			"service_integrations":   []any{},
			"service_name":           "my-mysql",
			"service_password":       "secret",
			"service_port":           12345,
			"service_type":           "mysql",
			"service_uri":            uri,
			"service_username":       "avnadmin",
			"state":                  "RUNNING",
			"static_ips":             nil,
			"tag":                    []any{},
			"tech_emails":            []any{},
			"termination_protection": false,
			"timeouts":               nil,
		},
		config: map[string]any{
			"cloud_name":        testCloud,
			"mysql_user_config": []any{mysqlUserConfig(nil)},
			"plan":              testPlan,
			"project":           testProject,
			"service_name":      "my-mysql",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{Component: "mysql", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
				{Component: "mysql", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypeReplica},
			},
			ConnectionInfo: &service.ConnectionInfoOut{
				Mysql: []string{uri},
				MysqlParams: []service.MysqlParamOut{{
					Dbname: "defaultdb", Host: host, Password: "secret", Port: "12345", SslMode: "REQUIRED", User: "avnadmin",
				}},
			},
			DiskSpaceMb: new(81920),
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-mysql",
			ServiceType: "mysql",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"dbname": "defaultdb", "host": host, "password": "secret", "port": "12345", "ssl-mode": "REQUIRED", "user": "avnadmin",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"backup_hour": 5, "backup_minute": 19,
				"ip_filter":     []any{"0.0.0.0/0", "::/0"},
				"mysql_version": "8.0.35",
				"mysql": map[string]any{
					"log_output":              "INSIGHTS",
					"long_query_time":         1,
					"slow_query_log":          true,
					"sql_mode":                "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES",
					"sql_require_primary_key": true,
				},
				"public_access": map[string]any{"mysql": true, "mysqlx": false, "prometheus": false},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 81920, DiskSpaceCapMb: new(409600), DiskSpaceStepMb: new(10240),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "mysql",
		},
	}
}

func mysqlUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"backup_hour": 5, "backup_minute": 19,
		"ip_filter":     []any{"0.0.0.0/0", "::/0"},
		"mysql_version": "8.0.35",
		"mysql": []any{map[string]any{
			"log_output": "INSIGHTS", "slow_query_log": true,
			"sql_mode":                "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES",
			"sql_require_primary_key": true,
		}},
		"public_access": []any{map[string]any{"mysql": true, "mysqlx": false, "prometheus": false}},
	}
	maps.Copy(result, options)
	return result
}

func mysqlUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions": nil,
		"admin_password":            "",
		"admin_username":            "",
		"backup_hour":               5,
		"backup_minute":             19,
		"binlog_retention_period":   0,
		"ip_filter":                 []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":          []any{},
		"ip_filter_string":          nil,
		"migration":                 []any{},
		"mysql": []any{map[string]any{
			"automatic_sp_privileges": false,
			"log_output":              "INSIGHTS",
			"long_query_time":         1,
			"slow_query_log":          true,
			"sql_mode":                "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES",
			"sql_require_primary_key": true,
		}},
		"mysql_incremental_backup": []any{},
		"mysql_version":            "8.0.35",
		"private_access":           []any{},
		"privatelink_access":       []any{},
		"project_to_fork_from":     "",
		"public_access": []any{map[string]any{
			"mysql": true, "mysqlx": false, "prometheus": false,
		}},
		"recovery_target_time": "",
		"service_log":          false,
		"service_to_fork_from": "",
		"static_ips":           false,
	}
	maps.Copy(result, options)
	return result
}
