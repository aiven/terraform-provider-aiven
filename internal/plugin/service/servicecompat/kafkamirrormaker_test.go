package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
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
		},
	})
}

func kafkaMirrorMakerFixture() fixture {
	const (
		host = "my-kafkamirrormaker.example.com"
		uri  = "https://avnadmin:secret@my-kafkamirrormaker.example.com"
	)

	return fixture{
		typeName: "aiven_kafka_mirrormaker",
		version:  1,
		state: map[string]any{
			"additional_disk_space":         "",
			"cloud_name":                    testCloud,
			"cmk_id":                        "",
			"components":                    []any{},
			"disk_space":                    nil,
			"disk_space_cap":                "0B",
			"disk_space_default":            "0B",
			"disk_space_step":               "0B",
			"disk_space_used":               "0B",
			"id":                            testProject + "/my-kafkamirrormaker",
			"kafka_mirrormaker_user_config": []any{kafkaMirrorMakerUserConfigState(nil)},
			"maintenance_window_dow":        "monday",
			"maintenance_window_enabled":    true,
			"maintenance_window_time":       "10:00:00",
			"plan":                          testPlan,
			"project":                       testProject,
			"project_vpc_id":                "",
			"service_host":                  host,
			"service_integrations":          []any{},
			"service_name":                  "my-kafkamirrormaker",
			"service_password":              nil,
			"service_port":                  0,
			"service_type":                  "kafka_mirrormaker",
			"service_uri":                   uri,
			"service_username":              nil,
			"state":                         "RUNNING",
			"static_ips":                    nil,
			"tag":                           []any{},
			"tech_emails":                   []any{},
			"termination_protection":        false,
			"timeouts":                      nil,
		},
		config: map[string]any{
			"cloud_name":                    testCloud,
			"kafka_mirrormaker_user_config": []any{kafkaMirrorMakerUserConfig(nil)},
			"plan":                          testPlan,
			"project":                       testProject,
			"service_name":                  "my-kafkamirrormaker",
		},
		api: &service.ServiceGetOut{
			CloudName:      testCloud,
			ConnectionInfo: &service.ConnectionInfoOut{},
			Maintenance:    &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:           testPlan,
			ServiceName:    "my-kafkamirrormaker",
			ServiceType:    "kafka_mirrormaker",
			ServiceUri:     uri,
			ServiceUriParams: map[string]string{
				"host": host,
				"port": "0",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
				"kafka_mirrormaker": map[string]any{
					"refresh_groups_interval_seconds": 600,
					"refresh_topics_enabled":          true,
					"refresh_topics_interval_seconds": 600,
					"tasks_max_per_cpu":               1,
				},
			},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb:     0,
			DiskSpaceCapMb:  new(0),
			DiskSpaceStepMb: new(0),
			NodeCount:       1,
			ServicePlan:     testPlan,
			ServiceType:     "kafka_mirrormaker",
		},
	}
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
