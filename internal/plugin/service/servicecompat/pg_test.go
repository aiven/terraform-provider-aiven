package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"
)

// TestPG also covers the logic all services share, e.g., the disks, static IPs, tags.
func TestPG(t *testing.T) {
	t.Parallel()
	runCompat(t, pgFixture, []compatCase{
		{
			name: "config matches the state",
		},
		{
			// SDKv2 stored the user config the API returned
			name:     "config has no user config",
			config:   map[string]any{"pg_user_config": nil},
			wantPlan: []string{"pg_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"pg_user_config": []any{map[string]any{
				"pg_version":  "16",
				"backup_hour": 3,
			}}},
			wantPlan: []string{"pg_user_config[0].pg", "pg_user_config[0].pglookout", "state"},
		},
		{
			// A refresh stored it
			name:  "static_ips refreshed to empty",
			state: map[string]any{"static_ips": []any{}},
		},
		{
			// Older versions stored it for no additional disk
			name:  "additional_disk_space 0B",
			state: map[string]any{"additional_disk_space": "0B"},
		},
		{
			name:   "additional_disk_space",
			state:  map[string]any{"additional_disk_space": "30GiB", "disk_space_used": "110GiB"},
			config: map[string]any{"additional_disk_space": "30GiB"},
			api:    func(s *service.ServiceGetOut) { s.DiskSpaceMb = new(112640) },
		},
		{
			name:   "project_vpc_id",
			state:  map[string]any{"project_vpc_id": testProject + "/" + testVPC},
			config: map[string]any{"project_vpc_id": testProject + "/" + testVPC},
			api:    func(s *service.ServiceGetOut) { s.ProjectVpcId = testVPC },
		},
		{
			name: "static_ips",
			state: map[string]any{
				"static_ips":     []any{"ip-1"},
				"pg_user_config": []any{pgUserConfigState(map[string]any{"static_ips": true})},
			},
			config: map[string]any{
				"static_ips":     []any{"ip-1"},
				"pg_user_config": []any{pgUserConfig(map[string]any{"static_ips": true})},
			},
			api: func(s *service.ServiceGetOut) { s.UserConfig["static_ips"] = true },
			ips: []staticip.StaticIPOut{{
				CloudName:         testCloud,
				ServiceName:       "my-pg",
				State:             staticip.StaticIPStateTypeAssigned,
				StaticIPAddressId: "ip-1",
			}},
		},
		{
			name:   "tag",
			state:  map[string]any{"tag": []any{map[string]any{"key": "env", "value": "dev"}}},
			config: map[string]any{"tag": []any{map[string]any{"key": "env", "value": "dev"}}},
			api:    func(s *service.ServiceGetOut) { s.Tags = map[string]string{"env": "dev"} },
		},
		{
			name:   "tech_emails",
			state:  map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			config: map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			api:    func(s *service.ServiceGetOut) { s.TechEmails = []service.TechEmailOut{{Email: "me@example.com"}} },
		},
		{
			name:   "termination_protection",
			state:  map[string]any{"termination_protection": true},
			config: map[string]any{"termination_protection": true},
			api:    func(s *service.ServiceGetOut) { s.TerminationProtection = true },
		},
		{
			name: "read replica",
			state: map[string]any{"service_integrations": []any{map[string]any{
				"integration_type":    "read_replica",
				"source_service_name": "my-source",
			}}},
			config: map[string]any{"service_integrations": []any{map[string]any{
				"integration_type":    "read_replica",
				"source_service_name": "my-source",
			}}},
			api: func(s *service.ServiceGetOut) {
				s.ServiceIntegrations = []service.ServiceIntegrationOut{{
					Active:          true,
					DestService:     new("my-pg"),
					Enabled:         true,
					IntegrationType: service.IntegrationTypeReadReplica,
					SourceService:   "my-source",
				}}
			},
		},
		{
			name:     "tag update",
			config:   map[string]any{"tag": []any{map[string]any{"key": "env", "value": "prod"}}},
			wantPlan: []string{"tag", "state"},
		},
		{
			// It changes the disks, the hosts, and the credentials
			name:   "plan update",
			config: map[string]any{"plan": "business-4"},
			wantPlan: []string{
				"additional_disk_space",
				"components",
				"disk_space_cap",
				"disk_space_default",
				"disk_space_step",
				"disk_space_used",
				"maintenance_window_enabled",
				"pg",
				"plan",
				"service_host",
				"service_password",
				"service_port",
				"service_uri",
				"service_username",
				"state",
			},
		},
	})
}

const testVPC = "00000000-0000-0000-0000-000000000002"

func pgFixture() fixture {
	const (
		host = "my-pg.example.com"
		uri  = "postgres://avnadmin:secret@my-pg.example.com:12345/defaultdb?sslmode=require"
	)

	return fixture{
		typeName: "aiven_pg",
		version:  1,
		state: map[string]any{
			"additional_disk_space": "",
			"cloud_name":            testCloud,
			"cmk_id":                "",
			"components": []any{map[string]any{
				"component":                   "pg",
				"connection_uri":              host + ":12345",
				"host":                        host,
				"kafka_authentication_method": "",
				"kafka_ssl_ca":                "",
				"port":                        12345,
				"privatelink_connection_id":   "",
				"route":                       "dynamic",
				"ssl":                         true,
				"usage":                       "primary",
			}},
			"disk_space":                 nil,
			"disk_space_cap":             "400GiB",
			"disk_space_default":         "80GiB",
			"disk_space_step":            "10GiB",
			"disk_space_used":            "80GiB",
			"id":                         testProject + "/my-pg",
			"maintenance_window_dow":     "monday",
			"maintenance_window_enabled": true,
			"maintenance_window_time":    "10:00:00",
			"pg": []any{map[string]any{
				"bouncer":         "",
				"dbname":          "defaultdb",
				"host":            host,
				"max_connections": 100,
				"params": []any{map[string]any{
					"database_name": "defaultdb",
					"host":          host,
					"password":      "secret",
					"port":          12345,
					"sslmode":       "require",
					"user":          "avnadmin",
				}},
				"password":     "secret",
				"port":         12345,
				"replica_uri":  "",
				"sslmode":      "require",
				"standby_uris": []any{},
				"syncing_uris": []any{},
				"uri":          uri,
				"uris":         []any{uri},
				"user":         "avnadmin",
			}},
			"pg_user_config":         []any{pgUserConfigState(nil)},
			"plan":                   testPlan,
			"project":                testProject,
			"project_vpc_id":         "",
			"service_host":           host,
			"service_integrations":   []any{},
			"service_name":           "my-pg",
			"service_password":       "secret",
			"service_port":           12345,
			"service_type":           "pg",
			"service_uri":            uri,
			"service_username":       "avnadmin",
			"state":                  "RUNNING",
			"static_ips":             nil,
			"tag":                    []any{},
			"tech_emails":            []any{},
			"termination_protection": false,
			"timeouts":               nil,
		},
		config: map[string]any{
			"cloud_name":     testCloud,
			"pg_user_config": []any{pgUserConfig(nil)},
			"plan":           testPlan,
			"project":        testProject,
			"service_name":   "my-pg",
		},
		api: &service.ServiceGetOut{
			CloudName: testCloud,
			Components: []service.ComponentOut{{
				Component: "pg",
				Host:      host,
				Port:      12345,
				Route:     service.RouteTypeDynamic,
				Ssl:       new(true),
				Usage:     service.UsageTypePrimary,
			}},
			ConnectionInfo: &service.ConnectionInfoOut{
				PG: []string{uri},
				PGParams: []service.PGParamOut{{
					Dbname:   "defaultdb",
					Host:     host,
					Password: "secret",
					Port:     "12345",
					Sslmode:  "require",
					User:     "avnadmin",
				}},
			},
			DiskSpaceMb: new(81920),
			Maintenance: &service.MaintenanceOut{Dow: service.MaintenanceDowTypeMonday, Time: "10:00:00", Enabled: true},
			Metadata:    map[string]any{"max_connections": 100},
			Plan:        testPlan,
			ServiceName: "my-pg",
			ServiceType: "pg",
			ServiceUri:  uri,
			ServiceUriParams: map[string]string{
				"dbname":   "defaultdb",
				"host":     host,
				"password": "secret",
				"port":     "12345",
				"sslmode":  "require",
				"user":     "avnadmin",
			},
			State: service.ServiceStateTypeRunning,
			// The options set, and the defaults
			UserConfig: map[string]any{
				"backup_hour":   3,
				"backup_minute": 15,
				"ip_filter":     []any{"0.0.0.0/0", "::/0"},
				"pg": map[string]any{
					"io_combine_limit":    16,
					"io_max_concurrency":  -1,
					"io_method":           "worker",
					"password_encryption": "scram-sha-256",
				},
				"pg_version": "16",
				"pglookout":  map[string]any{"max_failover_replication_time_lag": 60},
			},
			Users: []service.UserOut{{Username: "avnadmin", Password: "secret", Type: "primary"}},
		},
		specs: &project.ProjectServicePlanSpecsGetOut{
			DiskSpaceMb:     81920,
			DiskSpaceCapMb:  new(409600),
			DiskSpaceStepMb: new(10240),
			NodeCount:       1,
			ServicePlan:     testPlan,
			ServiceType:     "pg",
		},
	}
}

// pgUserConfig is the config that has every block SDKv2 stored.
func pgUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"pg_version": "16",
		"pg":         []any{map[string]any{"password_encryption": "scram-sha-256"}},
		"pglookout":  []any{map[string]any{"max_failover_replication_time_lag": 60}},
	}
	maps.Copy(result, options)
	return result
}

// pgUserConfigState is the user config SDKv2 stored: the API values, and zero values for the rest.
func pgUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"admin_password":        "",
		"admin_username":        "",
		"backup_hour":           3,
		"backup_interval_hours": 0,
		"backup_minute":         15,
		"enable_ipv6":           false,
		"ip_filter":             []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":      []any{},
		"ip_filter_string":      nil,
		"migration":             []any{},
		"pg": []any{map[string]any{
			"autovacuum_max_workers": 0,
			"io_combine_limit":       16,
			"io_max_concurrency":     -1,
			"io_method":              "worker",
			"jit":                    false,
			"log_line_prefix":        "",
			"password_encryption":    "scram-sha-256",
		}},
		"pg_read_replica": false,
		"pg_version":      "16",
		"pgbouncer":       []any{},
		"pglookout":       []any{map[string]any{"max_failover_replication_time_lag": 60}},
		"private_access":  []any{},
		"public_access":   []any{},
		"service_log":     false,
		"static_ips":      false,
		"timescaledb":     []any{},
		"variant":         "",
		"work_mem":        0,
	}
	maps.Copy(result, options)
	return result
}
