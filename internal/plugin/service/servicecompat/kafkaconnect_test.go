package servicecompat_test

import (
	"maps"
	"testing"

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

			// The API keeps the removed blocks
			wantApply: []string{
				"kafka_connect_user_config[0].kafka_connect",
				"kafka_connect_user_config[0].public_access",
			},
		},
	})
}

func kafkaConnectFixture() fixture {
	const (
		host = "my-kafkaconnect.example.com"
		uri  = "https://avnadmin:secret@my-kafkaconnect.example.com:443"
	)

	f := newFixture("kafka_connect", 1, "my-kafkaconnect", host, 443)
	zeroDisk(&f)
	maps.Copy(f.state, map[string]any{
		"components": []any{
			componentState("kafka_connect", host, 443, "primary"),
			componentState("kafka_connect", host, 443, "replica"),
		},
		"kafka_connect_user_config": []any{kafkaConnectUserConfigState(nil)},
		"service_uri":               uri,
	})
	f.config["kafka_connect_user_config"] = []any{kafkaConnectUserConfig(nil)}
	f.api.Components = []service.ComponentOut{
		componentAPI("kafka_connect", host, 443, service.UsageTypePrimary),
		componentAPI("kafka_connect", host, 443, service.UsageTypeReplica),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "443", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"ip_filter":     []any{"0.0.0.0/0", "::/0"},
		"kafka_connect": map[string]any{"consumer_isolation_level": "read_committed"},
		"public_access": map[string]any{"kafka_connect": true, "prometheus": false},
	}
	return f
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
