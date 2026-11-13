package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
)

func TestKafkaMirrorMaker(t *testing.T) {
	t.Parallel()
	runCompat(t, kafkaMirrorMakerFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"kafka_mirrormaker_user_config": nil},
			wantPlan: []string{"kafka_mirrormaker_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"kafka_mirrormaker_user_config": []any{map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
			}}},
			wantPlan: []string{
				"kafka_mirrormaker_user_config[0].kafka_mirrormaker",
				"state",
			},

			// The API keeps the removed blocks
			wantApply: []string{
				"kafka_mirrormaker_user_config[0].kafka_mirrormaker",
			},
		},
	})
}

func kafkaMirrorMakerFixture() fixture {
	const (
		host = "my-kafkamirrormaker.example.com"
		uri  = "https://avnadmin:secret@my-kafkamirrormaker.example.com"
	)

	f := newFixture("kafka_mirrormaker", 1, "my-kafkamirrormaker", host, 0)
	zeroDisk(&f)
	maps.Copy(f.state, map[string]any{
		"components":                    []any{},
		"kafka_mirrormaker_user_config": []any{kafkaMirrorMakerUserConfigState(nil)},
		"service_password":              nil,
		"service_uri":                   uri,
		"service_username":              nil,
	})
	f.config["kafka_mirrormaker_user_config"] = []any{kafkaMirrorMakerUserConfig(nil)}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "0"}
	f.api.UserConfig = map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
		"kafka_mirrormaker": map[string]any{
			"refresh_groups_interval_seconds": 600, "refresh_topics_enabled": true,
			"refresh_topics_interval_seconds": 600, "tasks_max_per_cpu": 1,
		},
	}
	f.api.Users = nil
	return f
}

func kafkaMirrorMakerUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
		"kafka_mirrormaker": []any{map[string]any{
			"refresh_topics_enabled":          true,
			"refresh_groups_interval_seconds": 600,
			"refresh_topics_interval_seconds": 600,
		}},
	}
	maps.Copy(result, options)
	return result
}

func kafkaMirrorMakerUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions": nil,
		"ip_filter":                 []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":          []any{},
		"ip_filter_string":          nil,
		"kafka_mirrormaker": []any{map[string]any{
			"admin_timeout_ms":                    0,
			"emit_checkpoints_enabled":            false,
			"emit_checkpoints_interval_seconds":   0,
			"groups":                              "",
			"groups_exclude":                      "",
			"offset_lag_max":                      0,
			"refresh_groups_enabled":              false,
			"refresh_groups_interval_seconds":     600,
			"refresh_topics_enabled":              true,
			"refresh_topics_interval_seconds":     600,
			"sync_group_offsets_enabled":          false,
			"sync_group_offsets_interval_seconds": 0,
			"sync_topic_configs_enabled":          false,
			"tasks_max_per_cpu":                   1,
		}},
		"preferred_zones":               nil,
		"sasl_oauthbearer_allowed_urls": nil,
		"service_log":                   false,
		"static_ips":                    false,
	}
	maps.Copy(result, options)
	return result
}
