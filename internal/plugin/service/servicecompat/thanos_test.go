package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestThanos(t *testing.T) {
	t.Parallel()
	runCompat(t, thanosFixture, []compatCase{
		{
			name: "config matches the state",
		},
		{
			name:     "config has no user config",
			config:   map[string]any{"thanos_user_config": nil},
			wantPlan: []string{"thanos_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"thanos_user_config": []any{map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
			}}},
			wantPlan: []string{
				"state",
				"thanos_user_config[0].public_access",
				"thanos_user_config[0].query_frontend",
			},
		},
		{
			// SDKv2 stored tech_emails as a list block; config uses the same shape
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

func thanosFixture() fixture {
	const (
		host   = "my-thanos.example.com"
		uri    = "https://avnadmin:secret@my-thanos.example.com:12345"
		query  = "https://my-thanos.example.com:12345/query"
		qfront = "https://my-thanos.example.com:12345/query-frontend"
		rw     = "https://my-thanos.example.com:12345/receiver/write"
	)

	return fixture{
		typeName: "aiven_thanos",
		version:  0,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{
				map[string]any{
					"component": "query", "connection_uri": host + ":12345", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12345, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
				map[string]any{
					"component": "receiver", "connection_uri": host + ":12348", "host": host,
					"kafka_authentication_method": "", "kafka_ssl_ca": "",
					"port": 12348, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
				},
			},
			"disk_space":                 nil,
			"disk_space_cap":             "1600GiB",
			"disk_space_default":         "160GiB",
			"disk_space_step":            "50GiB",
			"disk_space_used":            "160GiB",
			"id":                         testProject + "/my-thanos",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-thanos",
			"service_password":           "secret",
			"service_port":               12348,
			"service_type":               "thanos",
			"service_uri":                uri,
			"service_username":           "avnadmin",
			"state":                      "RUNNING",
			"static_ips":                 nil,
			"tag":                        []any{},
			"tech_emails":                []any{},
			"termination_protection":     false,
			"thanos": []any{map[string]any{
				"query_frontend_uri":        qfront,
				"query_uri":                 query,
				"receiver_remote_write_uri": rw,
				"uris":                      []any{uri},
			}},
			"thanos_user_config": []any{thanosUserConfigState(nil)},
			"timeouts":           nil,
		},
		config: map[string]any{
			"cloud_name":         testCloud,
			"thanos_user_config": []any{thanosUserConfig(nil)},
			"plan":               testPlan,
			"project":            testProject,
			"service_name":       "my-thanos",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{
				{Component: "query", Host: host, Port: 12345, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
				{Component: "receiver", Host: host, Port: 12348, Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary},
			},
			ConnectionInfo: &service.ConnectionInfoOut{
				Thanos:                 []string{uri},
				QueryUri:               new(query),
				QueryFrontendUri:       new(qfront),
				ReceiverRemoteWriteUri: new(rw),
			},
			DiskSpaceMb: new(163840),
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:        testPlan,
			ServiceName: "my-thanos",
			ServiceType: "thanos",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"host": host, "port": "12348", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
				"public_access": map[string]any{
					"query": true, "compactor": false, "query_frontend": false,
					"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
				},
				"query-frontend": map[string]any{"query-range.align-range-with-step": true},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 163840, DiskSpaceCapMb: new(1638400), DiskSpaceStepMb: new(51200),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "thanos",
		},
	}
}

func thanosUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"query_frontend": []any{map[string]any{"query_range_align_range_with_step": true}},
		"ip_filter":      []any{"0.0.0.0/0", "::/0"},
		"public_access": []any{map[string]any{
			"query": true, "compactor": false, "query_frontend": false,
			"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func thanosUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"compactor":        []any{},
		"env":              nil,
		"ip_filter":        []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object": []any{},
		"ip_filter_string": nil,
		"object_storage_usage_alert_threshold_gb": 0,
		"private_access":     []any{},
		"privatelink_access": []any{},
		"public_access": []any{map[string]any{
			"compactor": false, "query": true, "query_frontend": false,
			"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
		}},
		"query": []any{},
		"query_frontend": []any{map[string]any{
			"query_range_align_range_with_step": true,
		}},
		"receiver_ingesting": nil,
		"receiver_routing":   nil,
		"ruler":              nil,
		"service_log":        false,
		"static_ips":         false,
		"store":              nil,
	}
	maps.Copy(result, options)
	return result
}
