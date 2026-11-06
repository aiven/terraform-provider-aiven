package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestFlink(t *testing.T) {
	t.Parallel()
	runCompat(t, flinkFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"flink_user_config": nil},
			wantPlan: []string{"flink_user_config", "state"},
		},
		{
			name: "config has no ip_filter",
			config: map[string]any{"flink_user_config": []any{map[string]any{
				"flink_version":        "1.19",
				"number_of_task_slots": 10,
			}}},
		},
	})
}

func flinkFixture() fixture {
	const (
		host      = "my-flink.example.com"
		uri       = "https://avnadmin:secret@my-flink.example.com"
		hostPort1 = "my-flink.example.com:443"
		hostPort2 = "my-flink.example.com:8081"
	)

	return fixture{
		typeName: "aiven_flink",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{map[string]any{
				"component":                   "flink",
				"connection_uri":              host + ":443",
				"host":                        host,
				"kafka_authentication_method": "",
				"kafka_ssl_ca":                "",
				"port":                        443,
				"privatelink_connection_id":   "",
				"route":                       "dynamic",
				"ssl":                         true,
				"usage":                       "primary",
			}},
			"disk_space":         nil,
			"disk_space_cap":     "0B",
			"disk_space_default": "0B",
			"disk_space_step":    "0B",
			"disk_space_used":    "0B",
			"flink": []any{map[string]any{
				"host_ports": []any{hostPort1, hostPort2},
			}},
			"flink_user_config":          []any{flinkUserConfigState(nil)},
			"id":                         testProject + "/my-flink",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-flink",
			"service_password":           "secret",
			"service_port":               0,
			"service_type":               "flink",
			"service_uri":                uri,
			"service_username":           "avnadmin",
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"timeouts":                   nil,
		},
		config: map[string]any{
			"cloud_name":        testCloud,
			"flink_user_config": []any{flinkUserConfig(nil)},
			"plan":              testPlan,
			"project":           testProject,
			"service_name":      "my-flink",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{{
				Component: "flink",
				Host:      host,
				Port:      443,
				Route:     service.RouteTypeDynamic,
				Ssl:       new(true),
				Usage:     service.UsageTypePrimary,
			}},
			ConnectionInfo: &service.ConnectionInfoOut{
				Flink: []string{hostPort1, hostPort2},
			},
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-flink",
			ServiceType: "flink",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"host":     host,
				"port":     "0",
				"user":     "avnadmin",
				"password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"flink_version":        "1.19",
				"ip_filter":            []any{"0.0.0.0/0", "::/0"},
				"number_of_task_slots": 10,
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb:     0,
			DiskSpaceCapMb:  new(0),
			DiskSpaceStepMb: new(0),
			NodeCount:       1,
			ServicePlan:     testPlan,
			ServiceType:     "flink",
		},
	}
}

func flinkUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"flink_version":        "1.19",
		"ip_filter":            []any{"0.0.0.0/0", "::/0"},
		"number_of_task_slots": 10,
	}
	maps.Copy(result, options)
	return result
}

func flinkUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions": nil,
		"custom_code":               false,
		"flink_version":             "1.19",
		"ip_filter":                 []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":          []any{},
		"ip_filter_string":          nil,
		"number_of_task_slots":      10,
		"pekko_ask_timeout_s":       0,
		"pekko_framesize_b":         0,
		"privatelink_access":        []any{},
		"public_access":             []any{},
		"service_log":               false,
		"static_ips":                false,
	}
	maps.Copy(result, options)
	return result
}
