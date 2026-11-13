package servicecompat_test

import (
	"maps"
	"testing"

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

			// The API keeps the removed blocks
			wantApply: []string{
				"mysql_user_config[0].mysql",
				"mysql_user_config[0].public_access",
			},
		},
	})
}

func mysqlFixture() fixture {
	const (
		host = "my-mysql.example.com"
		uri  = "mysql://avnadmin:secret@my-mysql.example.com:12345/defaultdb?ssl-mode=REQUIRED"
	)

	f := newFixture("mysql", 1, "my-mysql", host, 12345)
	maps.Copy(f.state, map[string]any{
		"components": []any{
			componentState("mysql", host, 12345, "primary"),
			componentState("mysql", host, 12345, "replica"),
		},
		"mysql_user_config": []any{mysqlUserConfigState(nil)},
		"service_uri":       uri,
		"mysql": []any{map[string]any{
			"params": []any{map[string]any{
				"database_name": "defaultdb", "host": host, "password": "secret",
				"port": 12345, "sslmode": "REQUIRED", "user": "avnadmin",
			}},
			"replica_uri": "", "standby_uris": []any{}, "syncing_uris": []any{}, "uris": []any{uri},
		}},
	})
	f.config["mysql_user_config"] = []any{mysqlUserConfig(nil)}
	f.api.Components = []service.ComponentOut{
		componentAPI("mysql", host, 12345, service.UsageTypePrimary),
		componentAPI("mysql", host, 12345, service.UsageTypeReplica),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{
		Mysql: []string{uri},
		MysqlParams: []service.MysqlParamOut{{
			Dbname: "defaultdb", Host: host, Password: "secret", Port: "12345", SslMode: "REQUIRED", User: "avnadmin",
		}},
	}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{
		"dbname": "defaultdb", "host": host, "password": "secret", "port": "12345", "ssl-mode": "REQUIRED", "user": "avnadmin",
	}
	f.api.UserConfig = map[string]any{
		"backup_hour": 5, "backup_minute": 19,
		"ip_filter": []any{"0.0.0.0/0", "::/0"}, "mysql_version": "8.0.35",
		"mysql": map[string]any{
			"log_output": "INSIGHTS", "long_query_time": 1, "slow_query_log": true,
			"sql_mode":                "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES",
			"sql_require_primary_key": true,
		},
		"public_access": map[string]any{"mysql": true, "mysqlx": false, "prometheus": false},
	}
	return f
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
