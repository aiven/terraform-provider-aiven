package servicecompat_test

import (
	"maps"
	"testing"

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

	f := newFixture("clickhouse", 0, "my-clickhouse", host, 12345)
	maps.Copy(f.state, map[string]any{
		"clickhouse":             []any{map[string]any{"uris": []any{uri}}},
		"clickhouse_user_config": []any{clickhouseUserConfigState(nil)},
		"components":             []any{componentState("clickhouse", host, 12345, "primary")},
		"service_uri":            host + ":12345",
	})
	f.config["clickhouse_user_config"] = []any{clickhouseUserConfig(nil)}
	f.api.Components = []service.ComponentOut{componentAPI("clickhouse", host, 12345, service.UsageTypePrimary)}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{Clickhouse: []string{uri}}
	f.api.ServiceUri = host + ":12345"
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "12345", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"backup_hour": 3, "backup_minute": 30, "clickhouse_version": "25.8",
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
	}
	return f
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
