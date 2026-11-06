package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
)

func TestGrafana(t *testing.T) {
	t.Parallel()
	runCompat(t, grafanaFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"grafana_user_config": nil},
			wantPlan: []string{"grafana_user_config", "state"},
		},
	})
}

func grafanaFixture() fixture {
	const (
		host = "my-grafana.example.com"
		uri  = "https://avnadmin:secret@my-grafana.example.com:443"
	)

	return fixture{
		typeName: "aiven_grafana",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{map[string]any{
				"component": "grafana", "connection_uri": host + ":443", "host": host,
				"kafka_authentication_method": "", "kafka_ssl_ca": "",
				"port": 443, "privatelink_connection_id": "", "route": "dynamic", "ssl": true, "usage": "primary",
			}},
			"disk_space":         nil,
			"disk_space_cap":     "0B",
			"disk_space_default": "0B",
			"disk_space_step":    "0B",
			"disk_space_used":    "0B",
			"grafana": []any{map[string]any{
				"uris": []any{uri},
			}},
			"grafana_user_config":        []any{grafanaUserConfigState(nil)},
			"id":                         testProject + "/my-grafana",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"plan":                       testPlan,
			"project":                    testProject,
			"project_vpc_id":             "",
			"service_host":               host,
			"service_integrations":       []any{},
			"service_name":               "my-grafana",
			"service_password":           "secret",
			"service_port":               443,
			"service_type":               "grafana",
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
			"cloud_name":          testCloud,
			"grafana_user_config": []any{grafanaUserConfig(nil)},
			"plan":                testPlan,
			"project":             testProject,
			"service_name":        "my-grafana",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{{
				Component: "grafana", Host: host, Port: 443,
				Route: service.RouteTypeDynamic, Ssl: new(true), Usage: service.UsageTypePrimary,
			}},
			ConnectionInfo: &service.ConnectionInfoOut{Grafana: []string{uri}},
			Maintenance:    &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Plan:           testPlan,
			ServiceName:    "my-grafana",
			ServiceType:    "grafana",
			ServiceUri:     uri,
			ServiceUriParams: map[string]string{
				"host": host, "port": "443", "user": "avnadmin", "password": "secret",
			},
			State: service.ServiceStateTypeRunning,
			UserConfig: map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb: 0, DiskSpaceCapMb: new(0), DiskSpaceStepMb: new(0),
			NodeCount: 1, ServicePlan: testPlan, ServiceType: "grafana",
		},
	}
}

func grafanaUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
	}
	maps.Copy(result, options)
	return result
}

func grafanaUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":        nil,
		"alerting_enabled":                 false,
		"alerting_error_or_timeout":        "",
		"alerting_max_annotations_to_keep": 0,
		"alerting_nodata_or_nullvalues":    "",
		"allow_embedding":                  false,
		"auth_azuread":                     []any{},
		"auth_basic_enabled":               false,
		"auth_generic_oauth":               []any{},
		"auth_github":                      []any{},
		"auth_gitlab":                      []any{},
		"auth_google":                      []any{},
		"auth_jwt":                         []any{},
		"auth_okta":                        []any{},
		"auth_saml":                        []any{},
		"custom_domain":                    "",
		"enable_ipv6":                      false,
		"ip_filter":                        []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":                 []any{},
		"ip_filter_string":                 nil,
		"private_access":                   []any{},
		"privatelink_access":               []any{},
		"public_access":                    []any{},
		"service_log":                      false,
		"smtp":                             []any{},
		"static_ips":                       false,
		"unified_alerting":                 []any{},
	}
	maps.Copy(result, options)
	return result
}
