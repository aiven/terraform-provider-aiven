package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
)

func TestKafka(t *testing.T) {
	t.Parallel()
	runCompat(t, kafkaFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"kafka_user_config": nil},
			wantPlan: []string{"kafka_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"kafka_user_config": []any{map[string]any{
				"kafka_version":   "3.9",
				"schema_registry": true,
			}}},
			wantPlan: []string{
				"kafka_user_config[0].kafka_authentication_methods",
				"state",
			},

			// The API keeps the removed blocks
			wantApply: []string{
				"kafka_user_config[0].kafka_authentication_methods",
			},
		},
		{
			// The default, which the API doesn't have
			name:   "default_acl unset",
			state:  map[string]any{"default_acl": true},
			config: map[string]any{"default_acl": nil},
		},
	})
}

func kafkaFixture() fixture {
	const (
		host     = "my-kafka.example.com"
		port     = 12345
		portRest = 12348
		uri      = "kafka+ssl://my-kafka.example.com:12345"
		restURI  = "https://my-kafka.example.com:12348"
		srURI    = "https://my-kafka.example.com:12348"
	)

	f := newFixture("kafka", 1, "my-kafka", host, port)
	maps.Copy(f.state, map[string]any{
		"disk_space_cap": "0B", "disk_space_default": "90GiB", "disk_space_step": "0B", "disk_space_used": "90GiB",
		"default_acl": false,
		"components": []any{
			map[string]any{
				"component": "kafka", "connection_uri": host + ":12345", "host": host,
				"kafka_authentication_method": "certificate", "kafka_ssl_ca": "project_ca",
				"port": port, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
			},
			componentState("kafka_rest", host, portRest, "primary"),
		},
		"kafka":             []any{map[string]any{"access_cert": "CERT", "access_key": "KEY", "connect_uri": "", "rest_uri": restURI, "schema_registry_uri": srURI, "uris": []any{uri, uri}}},
		"kafka_user_config": []any{kafkaUserConfigState(nil)},
		"karapace":          nil,
		"service_password":  nil,
		"service_uri":       uri,
		"service_username":  nil,
	})
	f.config["default_acl"] = false
	f.config["kafka_user_config"] = []any{kafkaUserConfig(nil)}
	f.api.DiskSpaceMb = new(92160)
	f.api.Components = []service.ComponentOut{
		{
			Component: "kafka", Host: host, Port: port,
			KafkaAuthenticationMethod: "certificate", KafkaSslCa: "project_ca",
			Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary,
		},
		componentAPI("kafka_rest", host, portRest, service.UsageTypePrimary),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{
		Kafka: []string{uri, uri}, KafkaAccessCert: new("CERT"), KafkaAccessKey: new("KEY"),
		KafkaRestUri: new(restURI), SchemaRegistryUri: new(srURI),
	}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "12345"}
	f.api.UserConfig = map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"}, "kafka_version": "3.9", "schema_registry": true,
		"kafka_authentication_methods": map[string]any{"certificate": true, "sasl": false},
	}
	f.api.Users = nil
	f.specs.DiskSpaceMb = 92160
	f.specs.DiskSpaceCapMb = new(0)
	f.specs.DiskSpaceStepMb = new(0)
	return f
}

func kafkaUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter":       []any{"0.0.0.0/0", "::/0"},
		"kafka_version":   "3.9",
		"schema_registry": true,
		"kafka_authentication_methods": []any{map[string]any{
			"certificate": true,
			"sasl":        false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func kafkaUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":  nil,
		"aiven_kafka_topic_messages": false,
		"backup_interval_hours":      0,
		"backup_retention_days":      0,
		"custom_domain":              "",
		"enable_ipv6":                false,
		"follower_fetching":          []any{},
		"gcp_auth_allowed_urls":      nil,
		"inkless":                    []any{},
		"ip_filter":                  []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":           []any{},
		"ip_filter_string":           nil,
		"kafka":                      []any{},
		"kafka_authentication_methods": []any{map[string]any{
			"certificate": true,
			"sasl":        false,
		}},
		"kafka_connect":                  false,
		"kafka_connect_config":           []any{},
		"kafka_connect_plugin_versions":  []any{},
		"kafka_connect_secret_providers": []any{},
		"kafka_diskless":                 []any{},
		"kafka_rest":                     false,
		"kafka_rest_authorization":       false,
		"kafka_rest_config":              []any{},
		"kafka_sasl_mechanisms":          []any{},
		"kafka_version":                  "3.9",
		"karapace_version":               "",
		"letsencrypt_sasl":               false,
		"letsencrypt_sasl_privatelink":   false,
		"preferred_zones":                nil,
		"private_access":                 []any{},
		"privatelink_access":             []any{},
		"public_access":                  []any{},
		"sasl_oauthbearer_allowed_urls":  nil,
		"schema_registry":                true,
		"schema_registry_config":         []any{},
		"service_log":                    false,
		"single_zone":                    []any{},
		"static_ips":                     false,
		"tiered_storage":                 []any{},
	}
	maps.Copy(result, options)
	return result
}
