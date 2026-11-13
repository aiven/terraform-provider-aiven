package servicecompat_test

import (
	"maps"
	"testing"

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

			// The API keeps the removed blocks
			wantApply: []string{
				"valkey_user_config[0].public_access",
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

	f := newFixture("valkey", 0, "my-valkey", host, 12345)
	zeroDisk(&f)
	maps.Copy(f.state, map[string]any{
		"components": []any{
			componentState("valkey", host, 12345, "primary"),
			componentState("valkey", host, 12345, "replica"),
		},
		"valkey": []any{map[string]any{
			"password": "secret", "replica_uri": "", "slave_uris": []any{}, "uris": []any{uri},
		}},
		"valkey_user_config": []any{valkeyUserConfigState(nil)},
		"service_uri":        uri,
	})
	f.config["valkey_user_config"] = []any{valkeyUserConfig(nil)}
	f.api.Components = []service.ComponentOut{
		componentAPI("valkey", host, 12345, service.UsageTypePrimary),
		componentAPI("valkey", host, 12345, service.UsageTypeReplica),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{Valkey: []string{uri}, ValkeyPassword: new("secret")}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "12345", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"frequent_snapshots": true, "ip_filter": []any{"0.0.0.0/0", "::/0"},
		"valkey_active_defrag_ignore_bytes": 104857600, "valkey_active_defrag_threshold_lower": 10,
		"valkey_active_expire_effort": 1, "valkey_io_threads": 1, "valkey_lfu_decay_time": 1,
		"valkey_lfu_log_factor": 10, "valkey_maxmemory_policy": "noeviction", "valkey_ssl": true,
		"valkey_timeout": 300, "valkey_version": "7.2",
		"public_access": map[string]any{"valkey": true, "prometheus": false},
	}
	return f
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
