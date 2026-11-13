package servicecompat_test

import (
	"maps"
	"testing"

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

			// The API keeps the removed blocks
			wantApply: []string{
				"opensearch_user_config[0].index_patterns",
				"opensearch_user_config[0].opensearch_dashboards",
				"opensearch_user_config[0].public_access",
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

	f := newFixture("opensearch", 1, "my-opensearch", host, 12345)
	maps.Copy(f.state, map[string]any{
		"components": []any{
			componentState("opensearch", host, 12345, "primary"),
			componentState("opensearch", host, 12345, "replica"),
		},
		"opensearch": []any{map[string]any{
			"kibana_uri": "", "opensearch_dashboards_uri": dashURI,
			"password": "secret", "uris": []any{osURI}, "username": "avnadmin",
		}},
		"opensearch_user_config": []any{opensearchUserConfigState(nil)},
		"service_uri":            osURI,
	})
	f.config["opensearch_user_config"] = []any{opensearchUserConfig(nil)}
	f.api.Components = []service.ComponentOut{
		componentAPI("opensearch", host, 12345, service.UsageTypePrimary),
		componentAPI("opensearch", host, 12345, service.UsageTypeReplica),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{
		Opensearch: []string{osURI}, OpensearchDashboardsUri: new(dashURI),
		OpensearchUsername: new("avnadmin"), OpensearchPassword: new("secret"),
	}
	f.api.ServiceUri = osURI
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "12345", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"}, "opensearch_version": "2.17",
		"index_patterns": []any{
			map[string]any{"pattern": "logs-*", "max_index_count": 3, "sorting_algorithm": "alphabetical"},
		},
		"opensearch_dashboards": map[string]any{
			"enabled": true, "max_old_space_size": 128, "multiple_data_source_enabled": true,
			"opensearch_request_timeout": 30000, "session_keepalive": true, "session_ttl": "24h",
		},
		"public_access": map[string]any{"opensearch": true, "opensearch_dashboards": true, "prometheus": false},
	}
	return f
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
